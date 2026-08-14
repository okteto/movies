package catalog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// DefaultCopies is used for movies that don't declare how many copies exist.
const DefaultCopies = 3

// Movie mirrors the documents served by the catalog service.
type Movie struct {
	ID            json.Number `json:"id"`
	VoteAverage   float64     `json:"vote_average,omitempty"`
	OriginalTitle string      `json:"original_title,omitempty"`
	BackdropPath  string      `json:"backdrop_path,omitempty"`
	Price         float64     `json:"price,omitempty"`
	Overview      string      `json:"overview,omitempty"`
	Copies        *int        `json:"copies,omitempty"`
}

// TotalCopies returns the number of copies available for rental.
func (m Movie) TotalCopies() int {
	if m.Copies == nil || *m.Copies < 0 {
		return DefaultCopies
	}
	return *m.Copies
}

var client = &http.Client{Timeout: 5 * time.Second}

func url() string {
	if u := os.Getenv("CATALOG_URL"); u != "" {
		return u
	}
	return "http://catalog:8080/catalog"
}

// List returns the whole movie catalog.
func List() ([]Movie, error) {
	resp, err := client.Get(url())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("catalog returned %d", resp.StatusCode)
	}

	movies := []Movie{}
	if err := json.NewDecoder(resp.Body).Decode(&movies); err != nil {
		return nil, err
	}

	return movies, nil
}

// ByID indexes a catalog by movie id.
func ByID(movies []Movie) map[string]Movie {
	index := map[string]Movie{}
	for _, m := range movies {
		index[m.ID.String()] = m
	}
	return index
}
