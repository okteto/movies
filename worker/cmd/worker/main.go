package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strconv"

	"fmt"

	_ "github.com/lib/pq"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/Shopify/sarama"
	"github.com/okteto/movies/pkg/database"
	"github.com/okteto/movies/pkg/kafka"
)

var (
	topic             = kingpin.Flag("topic", "Topic name").Default("rentals").String()
	messageCountStart = kingpin.Flag("messageCountStart", "Message counter start from:").Int()
)

type RentalMessage struct {
	Email string  `json:"email"`
	Price float64 `json:"price"`
}

func main() {
	db := database.Open()
	defer db.Close()

	database.Ping(db)

	dropTableStmt := `DROP TABLE IF EXISTS rentals`
	if _, err := db.Exec(dropTableStmt); err != nil {
		log.Panic(err)
	}

	createTableStmt := `CREATE TABLE IF NOT EXISTS rentals (id VARCHAR(255) NOT NULL, email VARCHAR(255) NOT NULL, price VARCHAR(255) NOT NULL, PRIMARY KEY (id, email))`
	if _, err := db.Exec(createTableStmt); err != nil {
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
				catalogID := string(msg.Key)

				var rental RentalMessage
				if err := json.Unmarshal(msg.Value, &rental); err != nil {
					// Fallback for old message format (backwards compatibility)
					price, _ := strconv.ParseFloat(string(msg.Value), 64)
					rental.Price = price
					rental.Email = "unknown@example.com"
				}

				fmt.Printf("Received message: catalog_id %s email %s price %f\n", catalogID, rental.Email, rental.Price)
				insertDynStmt := `insert into "rentals"("id", "email", "price") values($1, $2, $3) on conflict(id, email) do update set price = $3`
				if _, err := db.Exec(insertDynStmt, catalogID, rental.Email, fmt.Sprintf("%f", rental.Price)); err != nil {
					log.Panic(err)
				}
			case msg := <-consumerReturns.Messages():
				catalogID := string(msg.Key)
				email := string(msg.Value)
				fmt.Printf("Received return message: catalogID %s email %s\n", catalogID, email)
				deleteStmt := `DELETE FROM rentals WHERE id = $1 AND email = $2`
				if _, err := db.Exec(deleteStmt, catalogID, email); err != nil {
					log.Panic(err)
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
