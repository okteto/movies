package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"

	"fmt"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/Shopify/sarama"
)

var (
	topic             = kingpin.Flag("topic", "Topic name").Default("rentals").String()
	messageCountStart = kingpin.Flag("messageCountStart", "Message counter start from:").Int()
)

// ConsumerGroupHandler implements sarama.ConsumerGroupHandler
type ConsumerGroupHandler struct {
	messageCount *int
}

func main() {
	kingpin.Parse()

	config := sarama.NewConfig()
	config.Version = sarama.V2_6_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	// Enable manual commit - we'll commit only after successful API calls
	config.Consumer.Offsets.AutoCommit.Enable = false

	consumerGroup, err := sarama.NewConsumerGroup([]string{"kafka:9092"}, "movies-worker-group", config)
	if err != nil {
		log.Panic(err)
	}
	defer consumerGroup.Close()

	handler := &ConsumerGroupHandler{
		messageCount: messageCountStart,
	}

	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			// Consume from both topics
			topics := []string{"rentals", "returns"}
			if err := consumerGroup.Consume(ctx, topics, handler); err != nil {
				log.Printf("Error from consumer: %v", err)
			}
			// Check if context was cancelled
			if ctx.Err() != nil {
				return
			}
		}
	}()

	<-signals
	fmt.Println("Interrupt is detected")
	cancel()
	wg.Wait()
	log.Println("Processed", *messageCountStart, "messages")
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages()
func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		*h.messageCount++

		// Determine message type based on topic
		if message.Topic == "rentals" {
			if !h.processRentalMessage(session, message) {
				// Don't commit if processing failed
				log.Printf("Failed to process rental message, will retry on next poll")
				continue
			}
		} else if message.Topic == "returns" {
			if !h.processReturnMessage(session, message) {
				// Don't commit if processing failed
				log.Printf("Failed to process return message, will retry on next poll")
				continue
			}
		}

		// Only mark message as consumed if processing was successful
		session.MarkMessage(message, "")
		// Commit the offset immediately after successful processing
		session.Commit()
	}
	return nil
}

// processRentalMessage handles rental messages and returns true if successful
func (h *ConsumerGroupHandler) processRentalMessage(session sarama.ConsumerGroupSession, message *sarama.ConsumerMessage) bool {
	fmt.Printf("Received message: movies %s price %s\n", string(message.Key), string(message.Value))
	price, _ := strconv.ParseFloat(string(message.Value), 64)

	// Extract baggage header from Kafka message
	var baggageHeader string
	for _, header := range message.Headers {
		if string(header.Key) == "baggage" {
			baggageHeader = string(header.Value)
			fmt.Printf("Baggage header found in Kafka message: %s\n", baggageHeader)
			break
		}
	}

	// Call API to create/update rental
	rentalData := map[string]string{
		"id":    string(message.Key),
		"price": fmt.Sprintf("%f", price),
	}
	jsonData, err := json.Marshal(rentalData)
	if err != nil {
		log.Printf("error marshaling rental data: %v", err)
		return false
	}

	// Create request with baggage header
	req, err := http.NewRequest("POST", "http://api:8080/internal/rentals", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("error creating rental request: %v", err)
		return false
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
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("API returned non-200 status: %d - message will not be committed", resp.StatusCode)
		return false
	}

	fmt.Printf("Successfully created/updated rental: %s - message committed\n", string(message.Key))
	return true
}

// processReturnMessage handles return messages and returns true if successful
func (h *ConsumerGroupHandler) processReturnMessage(session sarama.ConsumerGroupSession, message *sarama.ConsumerMessage) bool {
	catalogID := string(message.Value)
	fmt.Printf("Received return message: catalogID %s\n", catalogID)

	// Extract baggage header from Kafka message
	var baggageHeader string
	for _, header := range message.Headers {
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
		return false
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
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("API returned non-200 status: %d - message will not be committed", resp.StatusCode)
		return false
	}

	fmt.Printf("Successfully deleted rental: %s - message committed\n", catalogID)
	return true
}
