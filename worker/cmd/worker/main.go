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

// rentalMessage is the shared contract for the "rentals" Kafka topic value,
// produced by the rent (Java) service. It carries the price and the selected
// video quality tier.
type rentalMessage struct {
	Price string `json:"price"`
	Tier  string `json:"tier"`
}

// parseRentalMessage decodes the JSON payload. For robustness against any
// legacy plain-price messages still on the topic, it falls back to treating
// the whole value as the price with the default SD tier.
func parseRentalMessage(value []byte) rentalMessage {
	var msg rentalMessage
	if err := json.Unmarshal(value, &msg); err != nil || msg.Price == "" {
		return rentalMessage{Price: string(value), Tier: "SD"}
	}
	if msg.Tier == "" {
		msg.Tier = "SD"
	}
	return msg
}

func main() {
	db := database.Open()
	defer db.Close()

	database.Ping(db)

	dropTableStmt := `DROP TABLE IF EXISTS rentals`
	if _, err := db.Exec(dropTableStmt); err != nil {
		log.Panic(err)
	}

	createTableStmt := `CREATE TABLE IF NOT EXISTS rentals (id VARCHAR(255) NOT NULL UNIQUE, price VARCHAR(255) NOT NULL, tier VARCHAR(255) NOT NULL DEFAULT 'SD')`
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
				fmt.Printf("Received message: movies %s value %s\n", string(msg.Key), string(msg.Value))
				rental := parseRentalMessage(msg.Value)
				price, _ := strconv.ParseFloat(rental.Price, 64)
				insertDynStmt := `insert into "rentals"("id", "price", "tier") values($1, $2, $3) on conflict(id) do update set price = $2, tier = $3`
				if _, err := db.Exec(insertDynStmt, string(msg.Key), fmt.Sprintf("%f", price), rental.Tier); err != nil {
					log.Panic(err)
				}
			case msg := <-consumerReturns.Messages():
				catalogID := string(msg.Value)
				fmt.Printf("Received return message: catalogID %s\n", catalogID)
				deleteStmt := `DELETE FROM rentals WHERE id = $1`
				if _, err := db.Exec(deleteStmt, catalogID); err != nil {
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
