package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"log"

	"github.com/okteto/movies/pkg/database"

	"fmt"

	_ "github.com/lib/pq"
	"github.com/gorilla/mux"
)

var db *sql.DB

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
	Userid int
	Firstname string
	Lastname string
	Phone string
	City string
	State string
	Zip string
	Age int
	Gender string
}

func loadData() {
	// Create rentals table
	dropRentalsStmt := `DROP TABLE IF EXISTS rentals`
	if _, err := db.Exec(dropRentalsStmt); err != nil {
		log.Panic(err)
	}

	createRentalsStmt := `CREATE TABLE IF NOT EXISTS rentals (id VARCHAR(255) NOT NULL UNIQUE, price VARCHAR(255) NOT NULL)`
	if _, err := db.Exec(createRentalsStmt); err != nil {
		log.Panic(err)
	}

	// Create users table
	dropUsersStmt := `DROP TABLE IF EXISTS users`
	if _, err := db.Exec(dropUsersStmt); err != nil {
		log.Panic(err)
	}

	createUsersStmt := `CREATE TABLE IF NOT EXISTS users (user_id int NOT NULL UNIQUE, first_name varchar(255), last_name varchar(255), phone varchar(15), city varchar(255), state varchar(30), zip varchar(12), age int, gender varchar(10))`
	if _, err := db.Exec(createUsersStmt); err != nil {
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

	return
}

func handleRequests() {
	muxRouter := mux.NewRouter().StrictSlash(true)

	muxRouter.HandleFunc("/rentals", rentals)
	muxRouter.HandleFunc("/users", allUsers)
	muxRouter.HandleFunc("/users/{userid}", singleUser)
	muxRouter.HandleFunc("/catalog", catalogProxy).Methods("GET")
	muxRouter.HandleFunc("/rent", rentProxy).Methods("POST")
	muxRouter.HandleFunc("/rent/return", returnProxy).Methods("POST")

	// Internal API endpoints for worker service
	muxRouter.HandleFunc("/internal/rentals", createOrUpdateRental).Methods("POST")
	muxRouter.HandleFunc("/internal/rentals/{id}", deleteRental).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8080", muxRouter))
}

func rentals(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received request...")

	rows, err := db.Query("SELECT * FROM rentals")
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

	// Fetch catalog using the helper function with header propagation
	movies, err := fetchCatalog(r.Header)
	if err != nil {
		fmt.Println(err)
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

// Helper function to copy headers including baggage
func copyHeaders(dst http.Header, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

// fetchCatalog makes an internal request to the catalog service with header propagation
func fetchCatalog(headers http.Header) ([]Movie, error) {
	req, err := http.NewRequest("GET", "http://catalog:8080/catalog", nil)
	if err != nil {
		return nil, fmt.Errorf("error creating catalog request: %w", err)
	}

	// Propagate all headers including baggage, traceparent, tracestate, etc.
	copyHeaders(req.Header, headers)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling catalog service: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading catalog response: %w", err)
	}

	var movies []Movie
	if err := json.Unmarshal(body, &movies); err != nil {
		return nil, fmt.Errorf("error unmarshaling catalog: %w", err)
	}

	return movies, nil
}

// catalogProxy proxies requests to the catalog service
func catalogProxy(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Proxying request to catalog service...")

	// Create request to catalog service using internal Kubernetes service DNS
	req, err := http.NewRequest("GET", "http://catalog:8080/catalog", nil)
	if err != nil {
		fmt.Println("error creating catalog request", err)
		w.WriteHeader(500)
		return
	}

	// Propagate all headers including baggage, traceparent, tracestate, etc.
	copyHeaders(req.Header, r.Header)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error calling catalog service", err)
		w.WriteHeader(500)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Copy response status and body
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// rentProxy proxies rent requests to the rent service
func rentProxy(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Proxying request to rent service...")

	// Read the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("error reading request body", err)
		w.WriteHeader(500)
		return
	}
	defer r.Body.Close()

	// Create request to rent service using internal Kubernetes service DNS
	req, err := http.NewRequest("POST", "http://rent:8080/rent", bytes.NewReader(body))
	if err != nil {
		fmt.Println("error creating rent request", err)
		w.WriteHeader(500)
		return
	}

	// Propagate all headers including baggage, traceparent, tracestate, etc.
	copyHeaders(req.Header, r.Header)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error calling rent service", err)
		w.WriteHeader(500)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Copy response status and body
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// returnProxy proxies return requests to the rent service
func returnProxy(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Proxying request to rent service for return...")

	// Read the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("error reading request body", err)
		w.WriteHeader(500)
		return
	}
	defer r.Body.Close()

	// Create request to rent service using internal Kubernetes service DNS
	req, err := http.NewRequest("POST", "http://rent:8080/rent/return", bytes.NewReader(body))
	if err != nil {
		fmt.Println("error creating return request", err)
		w.WriteHeader(500)
		return
	}

	// Propagate all headers including baggage, traceparent, tracestate, etc.
	copyHeaders(req.Header, r.Header)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error calling rent service for return", err)
		w.WriteHeader(500)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Copy response status and body
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// RentalRequest represents the request body for creating/updating a rental
type RentalRequest struct {
	ID    string `json:"id"`
	Price string `json:"price"`
}

// createOrUpdateRental creates or updates a rental entry in the database
func createOrUpdateRental(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received internal request to create/update rental...")

	var rentalReq RentalRequest
	if err := json.NewDecoder(r.Body).Decode(&rentalReq); err != nil {
		fmt.Println("error decoding request body", err)
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	insertDynStmt := `insert into "rentals"("id", "price") values($1, $2) on conflict(id) do update set price = $2`
	if _, err := db.Exec(insertDynStmt, rentalReq.ID, rentalReq.Price); err != nil {
		fmt.Println("error inserting/updating rental", err)
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}

	fmt.Printf("Rental created/updated: ID=%s, Price=%s\n", rentalReq.ID, rentalReq.Price)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// deleteRental deletes a rental entry from the database
func deleteRental(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rentalID := vars["id"]

	fmt.Printf("Received internal request to delete rental: ID=%s\n", rentalID)

	deleteStmt := `DELETE FROM rentals WHERE id = $1`
	if _, err := db.Exec(deleteStmt, rentalID); err != nil {
		fmt.Println("error deleting rental", err)
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}

	fmt.Printf("Rental deleted: ID=%s\n", rentalID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
