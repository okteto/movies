package com.okteto.rent.controller;

import com.fasterxml.jackson.annotation.JsonProperty;
import org.apache.kafka.clients.producer.ProducerRecord;
import org.apache.kafka.common.header.internals.RecordHeader;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.kafka.support.SendResult;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestHeader;
import org.springframework.web.bind.annotation.RestController;

import java.nio.charset.StandardCharsets;
import java.util.LinkedList;
import java.util.List;
import java.util.Map;
import java.util.Collections;

@RestController
public class RentController {
    private static final String KAFKA_TOPIC_RENTALS = "rentals";
    private static final String KAFKA_TOPIC_RETURNS = "returns";

    private final Logger logger = LoggerFactory.getLogger(RentController.class);

    @Autowired
    private KafkaTemplate<String, String> kafkaTemplate;

    @GetMapping(path= "/rent", produces = "application/json")
    Map<String, String> healthz() {
            return Collections.singletonMap("status", "ok");
    }
    
    @PostMapping(path= "/rent", consumes = "application/json", produces = "application/json")
    List<String> rent(@RequestBody Rent rentInput,
                      @RequestHeader(value = "baggage", required = false) String baggage) {
        String catalogID = rentInput.getMovieID();
        Double price = rentInput.getPrice();

        logger.info("Rent [{},{}] received", catalogID, price);
        if (baggage != null) {
            logger.info("Baggage header received: {}", baggage);
        }

        // Create ProducerRecord to add custom headers
        ProducerRecord<String, String> record = new ProducerRecord<>(KAFKA_TOPIC_RENTALS, catalogID, price.toString());

        // Add baggage header to Kafka message if present
        if (baggage != null && !baggage.isEmpty()) {
            record.headers().add(new RecordHeader("baggage", baggage.getBytes(StandardCharsets.UTF_8)));
        }

        kafkaTemplate.send(record)
        .thenAccept(result -> logger.info("Message [{}] delivered with offset {}",
                        catalogID,
                        result.getRecordMetadata().offset()))
        .exceptionally(ex -> {
            logger.warn("Unable to deliver message [{}]. {}", catalogID, ex.getMessage());
            return null;
        });


        return new LinkedList<>();
    }

    @PostMapping(path= "/rent/return", consumes = "application/json", produces = "application/json")
    public Map<String, String> returnMovie(@RequestBody ReturnRequest returnRequest,
                                           @RequestHeader(value = "baggage", required = false) String baggage) {
        String catalogID = returnRequest.getMovieID();

        logger.info("Return [{}] received", catalogID);
        if (baggage != null) {
            logger.info("Baggage header received: {}", baggage);
        }

        // Create ProducerRecord to add custom headers
        ProducerRecord<String, String> record = new ProducerRecord<>(KAFKA_TOPIC_RETURNS, catalogID, catalogID);

        // Add baggage header to Kafka message if present
        if (baggage != null && !baggage.isEmpty()) {
            record.headers().add(new RecordHeader("baggage", baggage.getBytes(StandardCharsets.UTF_8)));
        }

        kafkaTemplate.send(record)
        .thenAccept(result -> logger.info("Return message [{}] delivered with offset {}",
                        catalogID,
                        result.getRecordMetadata().offset()))
        .exceptionally(ex -> {
            logger.warn("Unable to deliver return message [{}]. {}", catalogID, ex.getMessage());
            return null;
        });

        return Collections.singletonMap("status", "return processed");
    }

    public static class Rent {
        @JsonProperty("catalog_id")
        private String movieID;
        private Double price;

        public void setMovieID(String movieID) {
            this.movieID = movieID;
        }

        public String getMovieID() {
            return movieID;
        }


        public void setPrice(Double price) {
            this.price = price;
        }

        public Double getPrice() {
            return price;
        }
    }

    public static class ReturnRequest {
        @JsonProperty("catalog_id")
        private String movieID;

        public void setMovieID(String movieID) {
            this.movieID = movieID;
        }

        public String getMovieID() {
            return movieID;
        }
    }
}
