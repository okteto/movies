package catalog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

// DefaultCopies is used for movies that don't declare how many copies exist.
const DefaultCopies = 3

type Movie struct {
	ID     json.Number `json:"id"`
	Copies *int        `json:"copies"`
}

var client = &http.Client{Timeout: 5 * time.Second}

func url() string {
	if u := os.Getenv("CATALOG_URL"); u != "" {
		return u
	}
	return "http://catalog:8080/catalog"
}

// Copies returns how many copies of a movie the catalog holds.
func Copies(movieID string) (int, error) {
	resp, err := client.Get(url())
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("catalog returned %d", resp.StatusCode)
	}

	movies := []Movie{}
	if err := json.NewDecoder(resp.Body).Decode(&movies); err != nil {
		return 0, err
	}

	for _, m := range movies {
		if m.ID.String() != movieID {
			continue
		}
		if m.Copies == nil || *m.Copies < 0 {
			return DefaultCopies, nil
		}
		return *m.Copies, nil
	}

	return 0, fmt.Errorf("movie %s is not in the catalog", movieID)
}

// FormatPrice renders a price the way the rentals table stores it.
func FormatPrice(price float64) string {
	return strconv.FormatFloat(price, 'f', 2, 64)
}
