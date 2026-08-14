package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"os/signal"

	"fmt"

	_ "github.com/lib/pq"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/Shopify/sarama"
	"github.com/okteto/movies/pkg/catalog"
	"github.com/okteto/movies/pkg/database"
	"github.com/okteto/movies/pkg/kafka"
)

var (
	messageCountStart = kingpin.Flag("messageCountStart", "Message counter start from:").Int()
)

// RentMessage is the payload published by the rent service.
type RentMessage struct {
	UserEmail string  `json:"user_email"`
	MovieID   string  `json:"movie_id"`
	Price     float64 `json:"price"`
}

func main() {
	db := database.Open()
	defer db.Close()

	database.Ping(db)

	if err := database.EnsureSchema(db); err != nil {
		log.Panic(err)
	}

	master := kafka.GetMaster()
	defer master.Close()

	// Consumer for "rentals" topic
	consumerRentals, err := master.ConsumePartition("rentals", 0, sarama.OffsetNewest)
	if err != nil {
		log.Panic(err)
	}

	// Consumer for "returns" topic
	consumerReturns, err := master.ConsumePartition("returns", 0, sarama.OffsetNewest)
	if err != nil {
		log.Panic(err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	doneCh := make(chan struct{})

	go func() {
		for {
			select {
			case err := <-consumerRentals.Errors():
				fmt.Println(err)
			case msg := <-consumerRentals.Messages():
				*messageCountStart++
				if err := rent(db, msg.Value); err != nil {
					fmt.Println("rent rejected:", err)
				}
			case msg := <-consumerReturns.Messages():
				if err := returnMovie(db, msg.Value); err != nil {
					fmt.Println("return rejected:", err)
				}
			case <-signals:
				fmt.Println("Interrupt is detected")
				doneCh <- struct{}{}
			}
		}
	}()
	<-doneCh
	log.Println("Processed", *messageCountStart, "messages")
}

func rent(db *sql.DB, payload []byte) error {
	var m RentMessage
	if err := json.Unmarshal(payload, &m); err != nil {
		return fmt.Errorf("invalid rent message %q: %w", string(payload), err)
	}

	fmt.Printf("Received message: user %s movie %s price %f\n", m.UserEmail, m.MovieID, m.Price)

	if m.UserEmail == "" || m.MovieID == "" {
		return fmt.Errorf("rent message is missing the user or the movie")
	}

	copies, err := catalog.Copies(m.MovieID)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var banned bool
	if err := tx.QueryRow(`SELECT banned FROM users WHERE email = $1`, m.UserEmail).Scan(&banned); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user %s does not exist", m.UserEmail)
		}
		return err
	}
	if banned {
		return fmt.Errorf("user %s is banned", m.UserEmail)
	}

	var alreadyRented int
	if err := tx.QueryRow(
		`SELECT COUNT(*) FROM rentals WHERE movie_id = $1 AND user_email = $2 AND returned_at IS NULL`,
		m.MovieID, m.UserEmail,
	).Scan(&alreadyRented); err != nil {
		return err
	}
	if alreadyRented > 0 {
		return fmt.Errorf("user %s already rented movie %s", m.UserEmail, m.MovieID)
	}

	var active int
	if err := tx.QueryRow(
		`SELECT COUNT(*) FROM rentals WHERE movie_id = $1 AND returned_at IS NULL`,
		m.MovieID,
	).Scan(&active); err != nil {
		return err
	}
	if active >= copies {
		return fmt.Errorf("all %d copies of movie %s are rented", copies, m.MovieID)
	}

	if _, err := tx.Exec(
		`INSERT INTO rentals (user_email, movie_id, price) VALUES ($1, $2, $3)`,
		m.UserEmail, m.MovieID, catalog.FormatPrice(m.Price),
	); err != nil {
		return err
	}

	return tx.Commit()
}

func returnMovie(db *sql.DB, payload []byte) error {
	var m RentMessage
	if err := json.Unmarshal(payload, &m); err != nil {
		return fmt.Errorf("invalid return message %q: %w", string(payload), err)
	}

	fmt.Printf("Received return message: user %s movie %s\n", m.UserEmail, m.MovieID)

	result, err := db.Exec(
		`UPDATE rentals SET returned_at = NOW()
		 WHERE id = (
			SELECT id FROM rentals
			WHERE user_email = $1 AND movie_id = $2 AND returned_at IS NULL
			ORDER BY rented_at LIMIT 1
		 )`,
		m.UserEmail, m.MovieID,
	)
	if err != nil {
		return err
	}

	if rows, err := result.RowsAffected(); err == nil && rows == 0 {
		return fmt.Errorf("user %s has no active rental for movie %s", m.UserEmail, m.MovieID)
	}

	return nil
}
