package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/okteto/movies/pkg/database"

	"fmt"

	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	db = database.Open()
	defer db.Close()
	fmt.Println("Running server on port 8080...")
	http.HandleFunc("/rentals", rentals)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
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
