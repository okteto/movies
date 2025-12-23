package com.okteto.rent.controller;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.kafka.support.SendResult;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;

import java.util.LinkedList;
import java.util.List;
import java.util.Map;
import java.util.HashMap;
import java.util.Collections;

@RestController
public class RentController {
    private static final String KAFKA_TOPIC_RENTALS = "rentals";
    private static final String KAFKA_TOPIC_RETURNS = "returns";

    private final Logger logger = LoggerFactory.getLogger(RentController.class);

    @Autowired
    private KafkaTemplate<String, String> kafkaTemplate;

    private final ObjectMapper objectMapper = new ObjectMapper();

    @GetMapping(path= "/rent", produces = "application/json")
    Map<String, String> healthz() {
            return Collections.singletonMap("status", "ok");
    }
    
    @PostMapping(path= "/rent", consumes = "application/json", produces = "application/json")
    List<String> rent(@RequestBody Rent rentInput) {
        String catalogID = rentInput.getMovieID();
        Double price = rentInput.getPrice();

        // Get email from JWT authentication context (server-side, secure)
        Authentication authentication = SecurityContextHolder.getContext().getAuthentication();
        String email = authentication != null ? authentication.getName() : null;

        if (email == null || email.isEmpty()) {
            logger.error("No authenticated user found");
            return new LinkedList<>();
        }

        logger.info("Rent [{},{},{}] received", catalogID, price, email);

        try {
            Map<String, Object> rentalData = new HashMap<>();
            rentalData.put("email", email);
            rentalData.put("price", price);
            String messageValue = objectMapper.writeValueAsString(rentalData);

            kafkaTemplate.send(KAFKA_TOPIC_RENTALS, catalogID, messageValue)
            .thenAccept(result -> logger.info("Message [{}] delivered with offset {}",
                            catalogID,
                            result.getRecordMetadata().offset()))
            .exceptionally(ex -> {
                logger.warn("Unable to deliver message [{}]. {}", catalogID, ex.getMessage());
                return null;
            });
        } catch (Exception ex) {
            logger.error("Error serializing rental message", ex);
        }

        return new LinkedList<>();
    }

    @PostMapping(path= "/rent/return", consumes = "application/json", produces = "application/json")
    public Map<String, String> returnMovie(@RequestBody ReturnRequest returnRequest) {
        String catalogID = returnRequest.getMovieID();

        // Get email from JWT authentication context (server-side, secure)
        Authentication authentication = SecurityContextHolder.getContext().getAuthentication();
        String email = authentication != null ? authentication.getName() : null;

        if (email == null || email.isEmpty()) {
            logger.error("No authenticated user found");
            return Collections.singletonMap("status", "error: no authenticated user");
        }

        logger.info("Return [{},{}] received", catalogID, email);

        kafkaTemplate.send(KAFKA_TOPIC_RETURNS, catalogID, email)
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
