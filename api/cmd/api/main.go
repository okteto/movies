package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/okteto/movies/pkg/database"
	"github.com/okteto/movies/pkg/tracing"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var (
	db         *sql.DB
	httpClient = &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
)

func main() {
	ctx := context.Background()
	shutdown, err := tracing.Init(ctx, "api")
	if err != nil {
		log.Printf("tracing init failed: %v", err)
	} else {
		defer shutdown(ctx)
	}

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

	muxRouter.Handle("/rentals", otelhttp.NewHandler(http.HandlerFunc(rentals), "GET /rentals"))
	muxRouter.Handle("/users", otelhttp.NewHandler(http.HandlerFunc(allUsers), "GET /users"))
	muxRouter.Handle("/users/{userid}", otelhttp.NewHandler(http.HandlerFunc(singleUser), "GET /users/{userid}"))

	log.Fatal(http.ListenAndServe(":8080", muxRouter))
}

func rentals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tracer := otel.Tracer("api")

	ctx, dbSpan := tracer.Start(ctx, "db.query.rentals")
	rows, err := db.QueryContext(ctx, "SELECT * FROM rentals")
	if err != nil {
		dbSpan.RecordError(err)
		dbSpan.SetStatus(codes.Error, err.Error())
		dbSpan.End()
		fmt.Println("error listing rentals", err)
		w.WriteHeader(500)
		return
	}
	defer rows.Close()

	var rentalList []Rental
	for rows.Next() {
		var ren Rental
		if err := rows.Scan(&ren.Movie, &ren.Price); err != nil {
			dbSpan.RecordError(err)
			dbSpan.End()
			fmt.Println("error scanning row", err)
			os.Exit(1)
		}
		rentalList = append(rentalList, ren)
	}
	dbSpan.SetAttributes(attribute.Int("db.rows_returned", len(rentalList)))
	dbSpan.End()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://catalog:8080/catalog", nil)
	resp, err := httpClient.Do(req)
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
	for _, rental := range rentalList {
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
	ctx := r.Context()
	tracer := otel.Tracer("api")

	ctx, span := tracer.Start(ctx, "db.query.users")
	rows, err := db.QueryContext(ctx, "SELECT * FROM users")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End()
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
	span.SetAttributes(attribute.Int("db.rows_returned", len(users)))
	span.End()

	fmt.Println("Returned", len(users), "user records.")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func singleUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userid := vars["userid"]

	ctx := r.Context()
	tracer := otel.Tracer("api")

	ctx, span := tracer.Start(ctx, "db.query.user")
	span.SetAttributes(attribute.String("user.id", userid))
	row := db.QueryRowContext(ctx, "SELECT * FROM users WHERE user_id = $1", userid)

	var user User
	if err := row.Scan(&user.Userid, &user.Firstname, &user.Lastname, &user.Phone, &user.City, &user.State, &user.Zip, &user.Age, &user.Gender); err != nil {
		span.End()
		if err == sql.ErrNoRows {
			fmt.Println("No user was found")
			w.WriteHeader(404)
			return
		}
		log.Panic("error scanning returned user", err)
	}
	span.End()

	fmt.Println("Returned", user)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
