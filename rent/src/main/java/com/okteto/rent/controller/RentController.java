package com.okteto.rent.controller;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.kafka.support.SendResult;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;

import java.util.LinkedList;
import java.util.List;
import java.util.Map;
import java.util.Collections;

@RestController
public class RentController {
    private static final String KAFKA_TOPIC_RENTALS = "rentals";
    private static final String KAFKA_TOPIC_RETURNS = "returns";

    private final Logger logger = LoggerFactory.getLogger(RentController.class);
    private final ObjectMapper objectMapper = new ObjectMapper();

    @Autowired
    private KafkaTemplate<String, String> kafkaTemplate;

    @GetMapping(path= "/rent", produces = "application/json")
    Map<String, String> healthz() {
            return Collections.singletonMap("status", "ok");
    }
    
    @PostMapping(path= "/rent", consumes = "application/json", produces = "application/json")
    List<String> rent(@RequestBody Rent rentInput) {
        String catalogID = rentInput.getMovieID();
        Double price = rentInput.getPrice();
        String tier = rentInput.getTier();

        logger.info("Rent [{},{},{}] received", catalogID, price, tier);

        // The "rentals" message value carries both the price and the selected
        // tier as a small JSON payload. Producer (this service) and consumer
        // (the Go worker) share this contract.
        String payload = buildRentalPayload(price, tier);

        kafkaTemplate.send(KAFKA_TOPIC_RENTALS, catalogID, payload)
        .thenAccept(result -> logger.info("Message [{}] delivered with offset {}",
                        catalogID,
                        result.getRecordMetadata().offset()))
        .exceptionally(ex -> {
            logger.warn("Unable to deliver message [{}]. {}", catalogID, ex.getMessage());
            return null;
        });


        return new LinkedList<>();
    }

    private String buildRentalPayload(Double price, String tier) {
        ObjectNode node = objectMapper.createObjectNode();
        node.put("price", price == null ? "0" : price.toString());
        node.put("tier", Rent.normalizeTier(tier));
        try {
            return objectMapper.writeValueAsString(node);
        } catch (JsonProcessingException e) {
            // Should never happen for a plain object node; fall back defensively.
            logger.warn("Unable to serialize rental payload. {}", e.getMessage());
            return "{\"price\":\"" + (price == null ? "0" : price.toString())
                    + "\",\"tier\":\"" + Rent.normalizeTier(tier) + "\"}";
        }
    }

    @PostMapping(path= "/rent/return", consumes = "application/json", produces = "application/json")
    public Map<String, String> returnMovie(@RequestBody ReturnRequest returnRequest) {
        String catalogID = returnRequest.getMovieID();

        logger.info("Return [{}] received", catalogID);

        kafkaTemplate.send(KAFKA_TOPIC_RETURNS, catalogID, catalogID)
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
        static final String DEFAULT_TIER = "SD";

        @JsonProperty("catalog_id")
        private String movieID;
        private Double price;
        private String tier;

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

        public void setTier(String tier) {
            this.tier = tier;
        }

        public String getTier() {
            return tier;
        }

        // SD is the default; anything other than SD/HD falls back to SD.
        static String normalizeTier(String tier) {
            if (tier == null) {
                return DEFAULT_TIER;
            }
            String normalized = tier.trim().toUpperCase();
            return normalized.equals("HD") ? "HD" : DEFAULT_TIER;
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
