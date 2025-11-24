package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"fmt"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/Shopify/sarama"
	"github.com/okteto/movies/pkg/kafka"
)

var (
	topic             = kingpin.Flag("topic", "Topic name").Default("rentals").String()
	messageCountStart = kingpin.Flag("messageCountStart", "Message counter start from:").Int()
)

func main() {
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
				fmt.Printf("Received message: movies %s price %s\n", string(msg.Key), string(msg.Value))
				price, _ := strconv.ParseFloat(string(msg.Value), 64)

				// Extract baggage header from Kafka message
				var baggageHeader string
				for _, header := range msg.Headers {
					if string(header.Key) == "baggage" {
						baggageHeader = string(header.Value)
						fmt.Printf("Baggage header found in Kafka message: %s\n", baggageHeader)
						break
					}
				}

				// Call API to create/update rental
				rentalData := map[string]string{
					"id":    string(msg.Key),
					"price": fmt.Sprintf("%f", price),
				}
				jsonData, err := json.Marshal(rentalData)
				if err != nil {
					log.Printf("error marshaling rental data: %v", err)
					continue
				}

				// Create request with baggage header
				req, err := http.NewRequest("POST", "http://api:8080/internal/rentals", bytes.NewBuffer(jsonData))
				if err != nil {
					log.Printf("error creating rental request: %v", err)
					continue
				}
				req.Header.Set("Content-Type", "application/json")

				// Add baggage header to API call if present
				if baggageHeader != "" {
					req.Header.Set("baggage", baggageHeader)
					fmt.Printf("Propagating baggage header to API: %s\n", baggageHeader)
				}

				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					log.Printf("error calling API to create rental: %v", err)
					continue
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					log.Printf("API returned non-200 status: %d", resp.StatusCode)
				} else {
					fmt.Printf("Successfully created/updated rental: %s\n", string(msg.Key))
				}
			case msg := <-consumerReturns.Messages():
				catalogID := string(msg.Value)
				fmt.Printf("Received return message: catalogID %s\n", catalogID)

				// Extract baggage header from Kafka message
				var baggageHeader string
				for _, header := range msg.Headers {
					if string(header.Key) == "baggage" {
						baggageHeader = string(header.Value)
						fmt.Printf("Baggage header found in Kafka message: %s\n", baggageHeader)
						break
					}
				}

				// Call API to delete rental
				req, err := http.NewRequest("DELETE", fmt.Sprintf("http://api:8080/internal/rentals/%s", catalogID), nil)
				if err != nil {
					log.Printf("error creating delete request: %v", err)
					continue
				}

				// Add baggage header to API call if present
				if baggageHeader != "" {
					req.Header.Set("baggage", baggageHeader)
					fmt.Printf("Propagating baggage header to API: %s\n", baggageHeader)
				}

				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					log.Printf("error calling API to delete rental: %v", err)
					continue
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					log.Printf("API returned non-200 status: %d", resp.StatusCode)
				} else {
					fmt.Printf("Successfully deleted rental: %s\n", catalogID)
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
