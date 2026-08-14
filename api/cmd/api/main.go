package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/okteto/movies/pkg/catalog"
	"github.com/okteto/movies/pkg/database"
	"github.com/okteto/movies/pkg/session"

	"fmt"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var db *sql.DB

// demoUsers are seeded so the app has something to show on a fresh install.
var demoUsers = []User{
	{Email: "cindy@example.com", DisplayName: "Cindy Lopez"},
	{Email: "ramiro@example.com", DisplayName: "Ramiro Berrelleza"},
	{Email: "pablo@example.com", DisplayName: "Pablo Chico"},
}

type User struct {
	Email        string     `json:"email"`
	DisplayName  string     `json:"display_name"`
	Banned       bool       `json:"banned"`
	BanReason    string     `json:"ban_reason"`
	CreatedAt    time.Time  `json:"created_at"`
	ActiveCount  int        `json:"active_rentals"`
	TotalCount   int        `json:"total_rentals"`
	LastRentalAt *time.Time `json:"last_rental_at,omitempty"`
}

type Movie struct {
	ID            int     `json:"id"`
	VoteAverage   float64 `json:"vote_average,omitempty"`
	OriginalTitle string  `json:"original_title,omitempty"`
	BackdropPath  string  `json:"backdrop_path,omitempty"`
	Price         float64 `json:"price,omitempty"`
	Overview      string  `json:"overview,omitempty"`
	Copies        int     `json:"copies"`
	Available     int     `json:"available"`
	Rented        bool    `json:"rented"`
	RentalID      int     `json:"rental_id,omitempty"`
}

type Rental struct {
	ID         int        `json:"id"`
	UserEmail  string     `json:"user_email"`
	MovieID    string     `json:"movie_id"`
	Title      string     `json:"title"`
	Price      float64    `json:"price"`
	RentedAt   time.Time  `json:"rented_at"`
	ReturnedAt *time.Time `json:"returned_at,omitempty"`
}

type Redemption struct {
	ID        int       `json:"id"`
	UserEmail string    `json:"user_email"`
	GoodDeed  string    `json:"good_deed"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	db = database.Open()
	defer db.Close()

	if len(os.Args) > 1 && os.Args[1] == "load-data" {
		database.Ping(db)
		fmt.Println("Loading data...")
		loadData()
		return
	}

	fmt.Println("Running server on port 8080...")
	handleRequests()
}

func loadData() {
	if err := database.EnsureSchema(db); err != nil {
		log.Panic(err)
	}

	for _, u := range demoUsers {
		if _, err := db.Exec(
			`INSERT INTO users (email, display_name) VALUES ($1, $2) ON CONFLICT (email) DO NOTHING`,
			u.Email, u.DisplayName,
		); err != nil {
			log.Panic(err)
		}
	}

	fmt.Println("Seeded", len(demoUsers), "demo users")
}

func handleRequests() {
	r := mux.NewRouter().StrictSlash(true)

	r.HandleFunc("/auth/login", login).Methods(http.MethodPost)
	r.HandleFunc("/auth/logout", logout).Methods(http.MethodPost)
	r.HandleFunc("/auth/admin-login", adminLogin).Methods(http.MethodPost)
	r.HandleFunc("/auth/admin-logout", adminLogout).Methods(http.MethodPost)
	r.HandleFunc("/me", me).Methods(http.MethodGet)

	r.HandleFunc("/rentals", rentals).Methods(http.MethodGet)
	r.HandleFunc("/rentals/history", history).Methods(http.MethodGet)
	r.HandleFunc("/availability", availability).Methods(http.MethodGet)
	r.HandleFunc("/redemptions", createRedemption).Methods(http.MethodPost)

	r.HandleFunc("/adminapi/session", adminSession).Methods(http.MethodGet)
	r.HandleFunc("/adminapi/users", adminUsers).Methods(http.MethodGet)
	r.HandleFunc("/adminapi/users/{email}/rentals", adminUserRentals).Methods(http.MethodGet)
	r.HandleFunc("/adminapi/users/{email}/ban", adminBan).Methods(http.MethodPost)
	r.HandleFunc("/adminapi/users/{email}/unban", adminUnban).Methods(http.MethodPost)
	r.HandleFunc("/adminapi/redemptions", adminRedemptions).Methods(http.MethodGet)
	r.HandleFunc("/adminapi/redemptions/{id}/resolve", adminResolveRedemption).Methods(http.MethodPost)

	// Only reachable from inside the cluster: it is not exposed by the ingress.
	r.HandleFunc("/internal/rent-check", rentCheck).Methods(http.MethodGet)

	log.Fatal(http.ListenAndServe(":8080", r))
}

func writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		fmt.Println("error writing response", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func currentUser(w http.ResponseWriter, r *http.Request) (*User, bool) {
	email, err := session.User(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "you are not logged in")
		return nil, false
	}

	user, err := getUser(email)
	if err == sql.ErrNoRows {
		session.Clear(w, session.UserCookie)
		writeError(w, http.StatusUnauthorized, "you are not logged in")
		return nil, false
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return nil, false
	}

	return user, true
}

func requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	if session.IsAdmin(r) {
		return true
	}
	writeError(w, http.StatusUnauthorized, "admin credentials required")
	return false
}

func getUser(email string) (*User, error) {
	var u User
	err := db.QueryRow(
		`SELECT email, display_name, banned, ban_reason, created_at FROM users WHERE email = $1`,
		email,
	).Scan(&u.Email, &u.DisplayName, &u.Banned, &u.BanReason, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func displayNameFor(email string) string {
	name := strings.NewReplacer(".", " ", "_", " ", "-", " ").Replace(strings.Split(email, "@")[0])

	words := strings.Fields(name)
	for i, word := range words {
		words[i] = strings.ToUpper(word[:1]) + word[1:]
	}

	return strings.Join(words, " ")
}

func login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	email := strings.ToLower(strings.TrimSpace(body.Email))
	if !strings.Contains(email, "@") || strings.HasPrefix(email, "@") || strings.HasSuffix(email, "@") {
		writeError(w, http.StatusBadRequest, "a valid email is required")
		return
	}

	if _, err := db.Exec(
		`INSERT INTO users (email, display_name) VALUES ($1, $2) ON CONFLICT (email) DO NOTHING`,
		email, displayNameFor(email),
	); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	user, err := getUser(email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	session.Set(w, email)
	writeJSON(w, http.StatusOK, user)
}

func logout(w http.ResponseWriter, r *http.Request) {
	session.Clear(w, session.UserCookie)
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

func adminLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	username := os.Getenv("ADMIN_USERNAME")
	if username == "" {
		username = "admin"
	}
	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		password = "admin123"
	}

	if body.Username != username || body.Password != password {
		writeError(w, http.StatusUnauthorized, "wrong username or password")
		return
	}

	session.SetAdmin(w)
	writeJSON(w, http.StatusOK, map[string]bool{"admin": true})
}

func adminLogout(w http.ResponseWriter, r *http.Request) {
	session.Clear(w, session.AdminCookie)
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

func adminSession(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"admin": session.IsAdmin(r)})
}

func me(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, user)
}

// activeRentals returns the movies the user currently holds, indexed by movie id.
func activeRentals(email string) (map[string]Rental, error) {
	rows, err := db.Query(
		`SELECT id, movie_id, price, rented_at FROM rentals WHERE user_email = $1 AND returned_at IS NULL ORDER BY rented_at DESC`,
		email,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[string]Rental{}
	for rows.Next() {
		var rental Rental
		if err := rows.Scan(&rental.ID, &rental.MovieID, &rental.Price, &rental.RentedAt); err != nil {
			return nil, err
		}
		rental.UserEmail = email
		result[rental.MovieID] = rental
	}

	return result, rows.Err()
}

// activeByMovie counts how many copies of every movie are currently rented.
func activeByMovie() (map[string]int, error) {
	rows, err := db.Query(`SELECT movie_id, COUNT(*) FROM rentals WHERE returned_at IS NULL GROUP BY movie_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := map[string]int{}
	for rows.Next() {
		var movieID string
		var count int
		if err := rows.Scan(&movieID, &count); err != nil {
			return nil, err
		}
		counts[movieID] = count
	}

	return counts, rows.Err()
}

func rentals(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(w, r)
	if !ok {
		return
	}

	rented, err := activeRentals(user.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	movies, err := catalog.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	counts, err := activeByMovie()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result := []Movie{}
	for _, m := range movies {
		rental, isRented := rented[m.ID.String()]
		if !isRented {
			continue
		}
		movie := toMovie(m, counts[m.ID.String()])
		movie.Price = rental.Price
		movie.Rented = true
		movie.RentalID = rental.ID
		result = append(result, movie)
	}

	writeJSON(w, http.StatusOK, result)
}

func toMovie(m catalog.Movie, active int) Movie {
	id, _ := m.ID.Int64()
	copies := m.TotalCopies()
	available := copies - active
	if available < 0 {
		available = 0
	}

	return Movie{
		ID:            int(id),
		VoteAverage:   m.VoteAverage,
		OriginalTitle: m.OriginalTitle,
		BackdropPath:  m.BackdropPath,
		Price:         m.Price,
		Overview:      m.Overview,
		Copies:        copies,
		Available:     available,
	}
}

func history(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(w, r)
	if !ok {
		return
	}

	result, err := userHistory(user.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func userHistory(email string) ([]Rental, error) {
	rows, err := db.Query(
		`SELECT id, user_email, movie_id, price, rented_at, returned_at FROM rentals WHERE user_email = $1 ORDER BY rented_at DESC`,
		email,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	movies, err := catalog.List()
	if err != nil {
		return nil, err
	}
	index := catalog.ByID(movies)

	result := []Rental{}
	for rows.Next() {
		var rental Rental
		if err := rows.Scan(&rental.ID, &rental.UserEmail, &rental.MovieID, &rental.Price, &rental.RentedAt, &rental.ReturnedAt); err != nil {
			return nil, err
		}
		if movie, ok := index[rental.MovieID]; ok {
			rental.Title = movie.OriginalTitle
		} else {
			rental.Title = fmt.Sprintf("Movie %s (removed from catalog)", rental.MovieID)
		}
		result = append(result, rental)
	}

	return result, rows.Err()
}

func availability(w http.ResponseWriter, r *http.Request) {
	movies, err := catalog.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	counts, err := activeByMovie()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var rented map[string]Rental
	if email, err := session.User(r); err == nil {
		if rented, err = activeRentals(email); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	result := []Movie{}
	for _, m := range movies {
		movie := toMovie(m, counts[m.ID.String()])
		if rental, ok := rented[m.ID.String()]; ok {
			movie.Rented = true
			movie.RentalID = rental.ID
		}
		result = append(result, movie)
	}

	writeJSON(w, http.StatusOK, result)
}

func createRedemption(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(w, r)
	if !ok {
		return
	}

	var body struct {
		GoodDeed string `json:"good_deed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	deed := strings.TrimSpace(body.GoodDeed)
	if deed == "" {
		writeError(w, http.StatusBadRequest, "tell us about your good deed first")
		return
	}

	if !user.Banned {
		writeError(w, http.StatusBadRequest, "you are not banned")
		return
	}

	var pending int
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM redemptions WHERE user_email = $1 AND status = 'pending'`, user.Email,
	).Scan(&pending); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if pending > 0 {
		writeError(w, http.StatusConflict, "an admin is already reviewing your good deed")
		return
	}

	if _, err := db.Exec(
		`INSERT INTO redemptions (user_email, good_deed) VALUES ($1, $2)`, user.Email, deed,
	); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"status": "pending"})
}

func adminUsers(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	rows, err := db.Query(
		`SELECT u.email, u.display_name, u.banned, u.ban_reason, u.created_at,
			COUNT(r.id) FILTER (WHERE r.returned_at IS NULL) AS active,
			COUNT(r.id) AS total,
			MAX(r.rented_at) AS last_rental
		 FROM users u LEFT JOIN rentals r ON r.user_email = u.email
		 GROUP BY u.email ORDER BY u.created_at`,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.Email, &u.DisplayName, &u.Banned, &u.BanReason, &u.CreatedAt, &u.ActiveCount, &u.TotalCount, &u.LastRentalAt); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		users = append(users, u)
	}

	writeJSON(w, http.StatusOK, users)
}

func adminUserRentals(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	result, err := userHistory(mux.Vars(r)["email"])
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func adminBan(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	var body struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	reason := strings.TrimSpace(body.Reason)
	if reason == "" {
		reason = "misbehaving"
	}

	setBanned(w, mux.Vars(r)["email"], true, reason)
}

func adminUnban(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	setBanned(w, mux.Vars(r)["email"], false, "")
}

func setBanned(w http.ResponseWriter, email string, banned bool, reason string) {
	result, err := db.Exec(`UPDATE users SET banned = $1, ban_reason = $2 WHERE email = $3`, banned, reason, email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if rows, err := result.RowsAffected(); err == nil && rows == 0 {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	user, err := getUser(email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func adminRedemptions(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	rows, err := db.Query(`SELECT id, user_email, good_deed, status, created_at FROM redemptions ORDER BY created_at DESC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	redemptions := []Redemption{}
	for rows.Next() {
		var redemption Redemption
		if err := rows.Scan(&redemption.ID, &redemption.UserEmail, &redemption.GoodDeed, &redemption.Status, &redemption.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		redemptions = append(redemptions, redemption)
	}

	writeJSON(w, http.StatusOK, redemptions)
}

func adminResolveRedemption(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if body.Status != "approved" && body.Status != "rejected" {
		writeError(w, http.StatusBadRequest, "status must be approved or rejected")
		return
	}

	var email string
	if err := db.QueryRow(
		`UPDATE redemptions SET status = $1 WHERE id = $2 AND status = 'pending' RETURNING user_email`,
		body.Status, mux.Vars(r)["id"],
	).Scan(&email); err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "redemption request not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if body.Status == "approved" {
		if _, err := db.Exec(`UPDATE users SET banned = FALSE, ban_reason = '' WHERE email = $1`, email); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": body.Status, "user_email": email})
}

// rentCheck is called by the rent service before publishing to Kafka, so the
// user gets an immediate answer instead of a silently dropped message.
func rentCheck(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	movieID := r.URL.Query().Get("movie_id")
	action := r.URL.Query().Get("action")

	deny := func(reason string) {
		writeJSON(w, http.StatusOK, map[string]interface{}{"allowed": false, "reason": reason})
	}

	user, err := getUser(email)
	if err == sql.ErrNoRows {
		deny("you are not logged in")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if user.Banned {
		deny(fmt.Sprintf("you are banned: %s", user.BanReason))
		return
	}

	var active int
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM rentals WHERE movie_id = $1 AND user_email = $2 AND returned_at IS NULL`,
		movieID, email,
	).Scan(&active); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if action == "return" {
		if active == 0 {
			deny("you don't have this movie rented")
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"allowed": true})
		return
	}

	if active > 0 {
		deny("you already rented this movie")
		return
	}

	movies, err := catalog.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	movie, ok := catalog.ByID(movies)[movieID]
	if !ok {
		deny("this movie is not in the catalog anymore")
		return
	}

	counts, err := activeByMovie()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if counts[movieID] >= movie.TotalCopies() {
		deny("all copies of this movie are rented, try again later")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"allowed": true})
}
