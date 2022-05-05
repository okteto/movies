package main

import (
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

func main() {
	db := database.Open()
	defer db.Close()

	database.Ping(db)

	dropTableStmt := `DROP TABLE IF EXISTS rentals`
	if _, err := db.Exec(dropTableStmt); err != nil {
		log.Panic(err)
	}

	createTableStmt := `CREATE TABLE IF NOT EXISTS rentals (id VARCHAR(255) NOT NULL UNIQUE, price VARCHAR(255) NOT NULL)`
	if _, err := db.Exec(createTableStmt); err != nil {
		log.Panic(err)
	}

	master := kafka.GetMaster()
	defer master.Close()

	consumer, err := master.ConsumePartition(*topic, 0, sarama.OffsetOldest)
	if err != nil {
		log.Panic(err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	doneCh := make(chan struct{})
	go func() {
		for {
			select {
			case err := <-consumer.Errors():
				fmt.Println(err)
			case msg := <-consumer.Messages():
				*messageCountStart++
				fmt.Printf("Received message: movies %s price %s\n", string(msg.Key), string(msg.Value))
				price, _ := strconv.ParseFloat(string(msg.Value), 64)
				// price *= 0.7
				insertDynStmt := `insert into "rentals"("id", "price") values($1, $2) on conflict(id) do update set price = $2`
				if _, err := db.Exec(insertDynStmt, string(msg.Key), fmt.Sprintf("%f", price)); err != nil {
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
