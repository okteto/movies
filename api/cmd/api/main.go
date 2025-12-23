package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/okteto/movies/pkg/database"

	"fmt"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var db *sql.DB

// emailCache stores cached email lookups
type emailCache struct {
	mu      sync.RWMutex
	entries map[string]emailCacheEntry
}

type emailCacheEntry struct {
	email     string
	expiresAt time.Time
}

var cache = &emailCache{
	entries: make(map[string]emailCacheEntry),
}

// CustomClaims contains custom data we want from the token.
type CustomClaims struct {
	Scope                   string `json:"scope"`
	Email                   string `json:"email"`
	EmailVerified           bool   `json:"email_verified"`
	Name                    string `json:"name"`
	Nickname                string `json:"nickname"`
	Picture                 string `json:"picture"`
	UpdatedAt               string `json:"updated_at"`
	HttpsOktetoAuth0ComEmail string `json:"https://okteto.auth0.com/email"`
}

type contextKey string

const emailContextKey contextKey = "email"

// Validate does nothing for this example.
func (c CustomClaims) Validate(ctx context.Context) error {
	return nil
}

type UserInfo struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Nickname      string `json:"nickname"`
	Picture       string `json:"picture"`
	Sub           string `json:"sub"`
}

// get retrieves an email from cache if it exists and hasn't expired
func (c *emailCache) get(token string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[token]
	if !exists {
		return "", false
	}

	if time.Now().After(entry.expiresAt) {
		return "", false
	}

	return entry.email, true
}

// set stores an email in cache with a TTL
func (c *emailCache) set(token, email string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[token] = emailCacheEntry{
		email:     email,
		expiresAt: time.Now().Add(ttl),
	}
}

// cleanup removes expired entries from the cache
func (c *emailCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for token, entry := range c.entries {
		if now.After(entry.expiresAt) {
			delete(c.entries, token)
		}
	}
}

// getUserEmailFromAuth0 fetches user info from Auth0's userinfo endpoint
func getUserEmailFromAuth0(accessToken string) string {
	// Check cache first
	if email, found := cache.get(accessToken); found {
		return email
	}

	req, err := http.NewRequest("GET", "https://okteto.auth0.com/userinfo", nil)
	if err != nil {
		log.Printf("Error creating userinfo request: %v", err)
		return ""
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error fetching userinfo: %v", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Userinfo endpoint returned status: %d", resp.StatusCode)
		return ""
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		log.Printf("Error decoding userinfo: %v", err)
		return ""
	}

	// Cache the email for 5 minutes
	if userInfo.Email != "" {
		cache.set(accessToken, userInfo.Email, 5*time.Minute)
	}

	return userInfo.Email
}

// EnsureValidToken is a middleware that will check the validity of our JWT.
func EnsureValidToken() func(next http.Handler) http.Handler {
	issuerURL, err := url.Parse("https://okteto.auth0.com/")
	if err != nil {
		log.Fatalf("Failed to parse the issuer url: %v", err)
	}

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{"https://okteto.auth0.com/api/v2/"},
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &CustomClaims{}
			},
		),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		log.Fatalf("Failed to set up the jwt validator")
	}

	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		authHeader := r.Header.Get("Authorization")
		log.Printf("Encountered error while validating JWT: %v", err)
		log.Printf("Authorization header: %s", authHeader)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Failed to validate JWT."}`))
	}

	middleware := jwtmiddleware.New(
		jwtValidator.ValidateToken,
		jwtmiddleware.WithErrorHandler(errorHandler),
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// First validate the JWT
			middleware.CheckJWT(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				email := ""

				// Extract claims from the validated token
				claims, ok := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
				if ok && claims != nil {
					// Try to get email from custom claims in the token
					customClaims, ok := claims.CustomClaims.(*CustomClaims)
					if ok {
						// Try different possible email fields
						if customClaims.Email != "" {
							email = customClaims.Email
						} else if customClaims.HttpsOktetoAuth0ComEmail != "" {
							email = customClaims.HttpsOktetoAuth0ComEmail
						}
					}

					// If email not found in token, fetch from Auth0 userinfo endpoint
					if email == "" {
						// Get the access token from the Authorization header
						authHeader := r.Header.Get("Authorization")
						if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
							token := authHeader[7:]
							email = getUserEmailFromAuth0(token)
						}
					}

					// Last fallback to subject if no email found
					if email == "" && claims.RegisteredClaims.Subject != "" {
						email = claims.RegisteredClaims.Subject
					}
				}

				if email != "" {
					// Add email to context
					ctx := context.WithValue(r.Context(), emailContextKey, email)
					r = r.WithContext(ctx)
				}

				next.ServeHTTP(w, r)
			})).ServeHTTP(w, r)
		})
	}
}

func main() {
	db = database.Open()
	defer db.Close()

	// Start cache cleanup goroutine
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			cache.cleanup()
		}
	}()

	if len(os.Args) > 1 && os.Args[1] == "load-data" {
		database.Ping(db)
		fmt.Println("Loading data...")
		loadData()
		return
	}

	fmt.Println("Running server on port 8080...")
	handleRequests()
}

type Rental struct {
	Movie string
	Price string
}

type Movie struct {
	ID            int     `json:"id,omitempty"`
	VoteAverage   float64 `json:"vote_average,omitempty"`
	OriginalTitle string  `json:"original_title,omitempty"`
	BackdropPath  string  `json:"backdrop_path,omitempty"`
	Price         float64 `json:"price,omitempty"`
	Overview      string  `json:"overview,omitempty"`
}

type User struct {
	Userid    int
	Firstname string
	Lastname  string
	Phone     string
	City      string
	State     string
	Zip       string
	Age       int
	Gender    string
}

func loadData() {
	dropTableStmt := `DROP TABLE IF EXISTS users`
	if _, err := db.Exec(dropTableStmt); err != nil {
		log.Panic(err)
	}

	createTableStmt := `CREATE TABLE IF NOT EXISTS users (user_id int NOT NULL UNIQUE, first_name varchar(255), last_name varchar(255), phone varchar(15), city varchar(255), state varchar(30), zip varchar(12), age int, gender varchar(10))`
	if _, err := db.Exec(createTableStmt); err != nil {
		log.Panic(err)
	}

	jsonContent, err := os.ReadFile("data/users.json")
	if err != nil {
		log.Panic(err)
	}

	var users []User

	unmarshalErr := json.Unmarshal([]byte(jsonContent), &users)

	if unmarshalErr != nil {
		log.Panic(err)
	}

	for _, user := range users {
		insertStmt := `insert into "users"("user_id", "first_name", "last_name", "phone", "city", "state", "zip", "age", "gender") values($1, $2, $3, $4, $5, $6, $7, $8, $9)`
		if _, err := db.Exec(insertStmt, user.Userid, user.Firstname, user.Lastname, user.Phone, user.City, user.State, user.Zip, user.Age, user.Gender); err != nil {
			log.Panic(err)
		}
	}
}

func handleRequests() {
	muxRouter := mux.NewRouter().StrictSlash(true)

	// Apply JWT middleware to all routes
	jwtMiddleware := EnsureValidToken()

	muxRouter.Handle("/rentals", jwtMiddleware(http.HandlerFunc(rentals))).Methods("GET")
	muxRouter.Handle("/users", jwtMiddleware(http.HandlerFunc(allUsers))).Methods("GET")
	muxRouter.Handle("/users/{userid}", jwtMiddleware(http.HandlerFunc(singleUser))).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", muxRouter))
}

func rentals(w http.ResponseWriter, r *http.Request) {
	email, _ := r.Context().Value(emailContextKey).(string)
	log.Printf("Received rentals request from user: %s", email)

	rows, err := db.Query("SELECT id, price FROM rentals WHERE email = $1", email)
	if err != nil {
		fmt.Println("error listing rentals", err)
		w.WriteHeader(500)
		return
	}
	defer rows.Close()

	var rentals []Rental

	for rows.Next() {
		var r Rental
		if err := rows.Scan(&r.Movie, &r.Price); err != nil {
			fmt.Println("error scanning row", err)
			os.Exit(1)
		}
		rentals = append(rentals, r)
	}
	if err = rows.Err(); err != nil {
		fmt.Println("error in rows", err)
		os.Exit(1)
	}

	resp, err := http.Get("http://catalog:8080/catalog")
	if err != nil {
		fmt.Println("error listing catalog", err)
		w.WriteHeader(500)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error reading catalog", err)
		w.WriteHeader(500)
		return
	}

	movies := []Movie{}
	if err := json.Unmarshal(body, &movies); err != nil {
		fmt.Println("error unmarshaling catalog", err)
		w.WriteHeader(500)
		return
	}

	result := []Movie{}
	for _, rental := range rentals {
		for _, m := range movies {
			if rental.Movie == strconv.Itoa(m.ID) {
				price, _ := strconv.ParseFloat(rental.Price, 64)
				m.Price = price
				result = append(result, m)
			}
		}
	}

	fmt.Println("Returned", result)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func allUsers(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received request...")

	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		fmt.Println("error listing users", err)
		w.WriteHeader(500)
		return
	}
	defer rows.Close()

	var users []User

	for rows.Next() {
		var u User
		if err := rows.Scan(&u.Userid, &u.Firstname, &u.Lastname, &u.Phone, &u.City, &u.State, &u.Zip, &u.Age, &u.Gender); err != nil {
			log.Panic("error scanning row", err)
		}
		users = append(users, u)
	}
	if err = rows.Err(); err != nil {
		log.Panic("error in rows", err)
	}

	fmt.Println("Returned", len(users), "user records.")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func singleUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userid := vars["userid"]

	fmt.Println("Received request...")

	row := db.QueryRow("SELECT * FROM users WHERE user_id = $1", userid)

	var user User

	if err := row.Scan(&user.Userid, &user.Firstname, &user.Lastname, &user.Phone, &user.City, &user.State, &user.Zip, &user.Age, &user.Gender); err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("No user was found")
			w.WriteHeader(404)
			return
		} else {
			log.Panic("error scanning returned user", err)
		}
	}

	fmt.Println("Returned", user)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
