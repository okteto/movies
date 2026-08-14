package com.okteto.rent.controller;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.okteto.rent.api.RentCheckClient;
import com.okteto.rent.session.SessionCookie;
import jakarta.servlet.http.HttpServletRequest;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;

import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.Map;

@RestController
public class RentController {
    private static final String KAFKA_TOPIC_RENTALS = "rentals";
    private static final String KAFKA_TOPIC_RETURNS = "returns";

    private final Logger logger = LoggerFactory.getLogger(RentController.class);
    private final ObjectMapper mapper = new ObjectMapper();

    @Autowired
    private KafkaTemplate<String, String> kafkaTemplate;

    @Autowired
    private RentCheckClient rentCheck;

    @GetMapping(path = "/rent", produces = "application/json")
    Map<String, String> healthz() {
        return Collections.singletonMap("status", "ok");
    }

    @PostMapping(path = "/rent", consumes = "application/json", produces = "application/json")
    ResponseEntity<Map<String, String>> rent(HttpServletRequest request, @RequestBody Rent rentInput) {
        String email = SessionCookie.email(request);
        if (email == null) {
            return error(HttpStatus.UNAUTHORIZED, "you are not logged in");
        }

        String catalogID = rentInput.getMovieID();

        logger.info("Rent [{},{}] received", email, catalogID);

        RentCheckClient.Result check = rentCheck.check(email, catalogID, "rent");
        if (!check.allowed()) {
            return error(HttpStatus.CONFLICT, check.reason());
        }

        publish(KAFKA_TOPIC_RENTALS, catalogID, payload(email, catalogID));
        return ResponseEntity.accepted().body(Collections.singletonMap("status", "rental requested"));
    }

    @PostMapping(path = "/rent/return", consumes = "application/json", produces = "application/json")
    public ResponseEntity<Map<String, String>> returnMovie(HttpServletRequest request, @RequestBody ReturnRequest returnRequest) {
        String email = SessionCookie.email(request);
        if (email == null) {
            return error(HttpStatus.UNAUTHORIZED, "you are not logged in");
        }

        String catalogID = returnRequest.getMovieID();

        logger.info("Return [{},{}] received", email, catalogID);

        RentCheckClient.Result check = rentCheck.check(email, catalogID, "return");
        if (!check.allowed()) {
            return error(HttpStatus.CONFLICT, check.reason());
        }

        publish(KAFKA_TOPIC_RETURNS, catalogID, payload(email, catalogID));
        return ResponseEntity.accepted().body(Collections.singletonMap("status", "return requested"));
    }

    private void publish(String topic, String key, String message) {
        kafkaTemplate.send(topic, key, message)
                .thenAccept(result -> logger.info("Message [{}] delivered to {} with offset {}",
                        message,
                        topic,
                        result.getRecordMetadata().offset()))
                .exceptionally(ex -> {
                    logger.warn("Unable to deliver message [{}]. {}", message, ex.getMessage());
                    return null;
                });
    }

    // the price is not part of the message: the worker charges what the catalog says
    private String payload(String email, String catalogID) {
        Map<String, Object> message = new LinkedHashMap<>();
        message.put("user_email", email);
        message.put("movie_id", catalogID);

        try {
            return mapper.writeValueAsString(message);
        } catch (Exception e) {
            throw new IllegalStateException("unable to serialize the kafka message", e);
        }
    }

    private ResponseEntity<Map<String, String>> error(HttpStatus status, String reason) {
        return ResponseEntity.status(status).body(Collections.singletonMap("error", reason));
    }

    public static class Rent {
        @JsonProperty("catalog_id")
        private String movieID;

        public void setMovieID(String movieID) {
            this.movieID = movieID;
        }

        public String getMovieID() {
            return movieID;
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
