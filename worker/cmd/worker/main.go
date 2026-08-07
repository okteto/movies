package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/Shopify/sarama"
	_ "github.com/lib/pq"
	"github.com/okteto/movies/pkg/database"
	"github.com/okteto/movies/pkg/kafka"
	"github.com/okteto/movies/pkg/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	topic             = kingpin.Flag("topic", "Topic name").Default("rentals").String()
	messageCountStart = kingpin.Flag("messageCountStart", "Message counter start from:").Int()
)

// kafkaHeaderCarrier extracts W3C trace context from Sarama message headers.
type kafkaHeaderCarrier []*sarama.RecordHeader

func (c kafkaHeaderCarrier) Get(key string) string {
	for _, h := range c {
		if string(h.Key) == key {
			return string(h.Value)
		}
	}
	return ""
}

func (c kafkaHeaderCarrier) Set(key, val string) {}

func (c kafkaHeaderCarrier) Keys() []string {
	keys := make([]string, len(c))
	for i, h := range c {
		keys[i] = string(h.Key)
	}
	return keys
}

func main() {
	ctx := context.Background()
	shutdown, err := tracing.Init(ctx, "worker")
	if err != nil {
		log.Printf("tracing init failed: %v", err)
	} else {
		defer shutdown(ctx)
	}

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

	consumerRentals, err := master.ConsumePartition("rentals", 0, sarama.OffsetNewest)
	if err != nil {
		log.Panic(err)
	}

	consumerReturns, err := master.ConsumePartition("returns", 0, sarama.OffsetNewest)
	if err != nil {
		log.Panic(err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	doneCh := make(chan struct{})

	tracer := otel.Tracer("worker")

	go func() {
		for {
			select {
			case err := <-consumerRentals.Errors():
				fmt.Println(err)

			case msg := <-consumerRentals.Messages():
				*messageCountStart++
				movieID := string(msg.Key)
				price, _ := strconv.ParseFloat(string(msg.Value), 64)
				fmt.Printf("Received rental message: movie=%s price=%s\n", movieID, msg.Value)

				parentCtx := otel.GetTextMapPropagator().Extract(
					context.Background(),
					propagation.MapCarrier(headersToMap(msg.Headers)),
				)
				msgCtx, span := tracer.Start(parentCtx, "worker.process_rental",
					trace.WithSpanKind(trace.SpanKindConsumer),
					trace.WithAttributes(
						attribute.String("kafka.topic", "rentals"),
						attribute.String("movie.id", movieID),
						attribute.Float64("movie.price", price),
					),
				)

				insertStmt := `insert into "rentals"("id", "price") values($1, $2) on conflict(id) do update set price = $2`
				_, dbErr := db.ExecContext(msgCtx, insertStmt, movieID, fmt.Sprintf("%f", price))
				if dbErr != nil {
					span.RecordError(dbErr)
					span.SetStatus(codes.Error, dbErr.Error())
					span.End()
					log.Panic(dbErr)
				}
				span.End()

			case msg := <-consumerReturns.Messages():
				catalogID := string(msg.Value)
				fmt.Printf("Received return message: catalogID=%s\n", catalogID)

				parentCtx := otel.GetTextMapPropagator().Extract(
					context.Background(),
					propagation.MapCarrier(headersToMap(msg.Headers)),
				)
				_, span := tracer.Start(parentCtx, "worker.process_return",
					trace.WithSpanKind(trace.SpanKindConsumer),
					trace.WithAttributes(
						attribute.String("kafka.topic", "returns"),
						attribute.String("movie.id", catalogID),
					),
				)

				deleteStmt := `DELETE FROM rentals WHERE id = $1`
				if _, err := db.Exec(deleteStmt, catalogID); err != nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, err.Error())
					span.End()
					log.Panic(err)
				}
				span.End()

			case <-signals:
				fmt.Println("Interrupt is detected")
				doneCh <- struct{}{}
			}
		}
	}()
	<-doneCh
	log.Println("Processed", *messageCountStart, "messages")
}

func headersToMap(headers []*sarama.RecordHeader) map[string]string {
	m := make(map[string]string, len(headers))
	for _, h := range headers {
		if h != nil {
			m[string(h.Key)] = string(h.Value)
		}
	}
	return m
}
