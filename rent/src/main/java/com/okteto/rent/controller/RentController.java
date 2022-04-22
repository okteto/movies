package com.okteto.rent.controller;

import com.fasterxml.jackson.annotation.JsonProperty;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.kafka.support.SendResult;
import org.springframework.util.concurrent.ListenableFuture;
import org.springframework.util.concurrent.ListenableFutureCallback;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;

import javax.servlet.http.HttpServletResponse;
import java.util.LinkedList;
import java.util.List;

@RestController
public class RentController {
    private static final String KAFKA_TOPIC = "rentals";

    private final Logger logger = LoggerFactory.getLogger(RentController.class);

    @Autowired
    private KafkaTemplate<String, String> kafkaTemplate;

    @PostMapping(path= "/rent", consumes = "application/json", produces = "application/json")
    List<String> rent(@RequestBody Rent rentInput, HttpServletResponse response) {
        String catalogID = rentInput.getMovieID();
        String price = rentInput.getPrice();

        logger.info("Rent [{},{}] received", catalogID, price);

        ListenableFuture<SendResult<String, String>> future = kafkaTemplate.send(KAFKA_TOPIC, catalogID, price);

        future.addCallback(new ListenableFutureCallback<SendResult<String, String>>() {
            @Override
            public void onSuccess(SendResult<String, String> result) {
                logger.info("Message [{}] delivered with offset {}",
                        catalogID,
                        result.getRecordMetadata().offset());
            }

            @Override
            public void onFailure(Throwable ex) {
                logger.warn("Unable to deliver message [{}]. {}",
                        catalogID,
                        ex.getMessage());
            }
        });

        return new LinkedList<>();
    }

    public static class Rent {
        @JsonProperty("catalog_id")
        private String movieID;
        private String price;

        public void setMovieID(String movieID) {
            this.movieID = movieID;
        }

        public String getMovieID() {
            return movieID;
        }


        public void setPrice(String price) {
            this.price = price;
        }

        public String getPrice() {
            return price;
        }
    }
}
