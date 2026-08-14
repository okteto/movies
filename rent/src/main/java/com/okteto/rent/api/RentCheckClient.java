package com.okteto.rent.api;

import com.fasterxml.jackson.databind.DeserializationFeature;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.springframework.stereotype.Component;

import java.net.URI;
import java.net.URLEncoder;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.charset.StandardCharsets;
import java.time.Duration;

/**
 * Asks the api service whether a rent or return is allowed, so the user gets an
 * immediate answer instead of a message that the worker silently drops.
 */
@Component
public class RentCheckClient {
    private final HttpClient client = HttpClient.newBuilder().connectTimeout(Duration.ofSeconds(5)).build();
    private final ObjectMapper mapper = new ObjectMapper()
            .configure(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES, false);
    private final String apiUrl;

    public RentCheckClient() {
        String url = System.getenv("API_URL");
        this.apiUrl = (url == null || url.isEmpty()) ? "http://api:8080" : url;
    }

    public record Result(boolean allowed, String reason) {
    }

    public Result check(String email, String movieID, String action) {
        String url = String.format("%s/internal/rent-check?email=%s&movie_id=%s&action=%s",
                apiUrl, encode(email), encode(movieID), encode(action));

        try {
            HttpRequest request = HttpRequest.newBuilder(URI.create(url))
                    .timeout(Duration.ofSeconds(5))
                    .GET()
                    .build();

            HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());
            if (response.statusCode() != 200) {
                return new Result(false, "the rental service is not available, try again later");
            }

            return mapper.readValue(response.body(), Result.class);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            return new Result(false, "the rental service is not available, try again later");
        } catch (Exception e) {
            return new Result(false, "the rental service is not available, try again later");
        }
    }

    private static String encode(String value) {
        return URLEncoder.encode(value == null ? "" : value, StandardCharsets.UTF_8);
    }
}
