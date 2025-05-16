package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/okteto/movies/pkg/database"

	"fmt"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Rental struct {
	Movie string
	Price string
}

type RentalHistory struct {
	ID        int
	MovieID   string
	Price     string
	CreatedAt string
	Title     string
}

type Movie struct {
	ID            int     `json:"id,omitempty"`
	VoteAverage   float64 `json:"vote_average,omitempty"`
	OriginalTitle string  `json:"original_title,omitempty"`
	BackdropPath  string  `json:"backdrop_path,omitempty"`
	Price         float64 `json:"price,omitempty"`
	Overview      string  `json:"overview,omitempty"`
}

var db *sql.DB



func main() {
	db = database.Open()
	defer db.Close()

	fmt.Println("Running server on port 8080...")
	
	muxRouter := mux.NewRouter().StrictSlash(true)

	muxRouter.HandleFunc("/rentals", rentals)
	muxRouter.HandleFunc("/rentals-history", rentalsHistory)
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

func rentalsHistory(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received request...")

	rows, err := db.Query("SELECT id, movie_id, price, rented_at FROM rentals_history WHERE id = '10000000'")
	if err != nil {
		fmt.Println("error listing rentals history", err)
		w.WriteHeader(500)
		return
	}
	defer rows.Close()

	var rentalsHistory []RentalHistory

	for rows.Next() {
		var r RentalHistory
		if err := rows.Scan(&r.ID, &r.MovieID, &r.Price, &r.CreatedAt); err != nil {
			fmt.Println("error scanning row", err)
			os.Exit(1)
		}
		rentalsHistory = append(rentalsHistory, r)
	}
	if err = rows.Err(); err != nil {
		fmt.Println("error in rows", err)
		os.Exit(1)
	}

	if (rentalsHistory == nil  || len(rentalsHistory) == 0) {
		fmt.Println("no history found")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]RentalHistory{})
		return
	}

	resp, err := http.Get("http://catalog:8080/catalog")
	if err != nil {
		fmt.Println("error listing catalog", err)
		w.WriteHeader(500)
		return
	}

	body, err := io.ReadAll(resp.Body)
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

	movieTitles := make(map[int]string)
	for _, m := range movies {
		movieTitles[m.ID] = m.OriginalTitle
	}

	for i, _ := range rentalsHistory {
		movieId, err := strconv.Atoi(rentalsHistory[i].MovieID)
		if err != nil {
			continue
		}

		rentalsHistory[i].Title = movieTitles[movieId]
	}

	fmt.Println("Returned", rentalsHistory)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rentalsHistory)
}
