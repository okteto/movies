package com.okteto.rent.security;

import com.auth0.jwk.Jwk;
import com.auth0.jwk.JwkProvider;
import com.auth0.jwk.JwkProviderBuilder;
import com.auth0.jwt.JWT;
import com.auth0.jwt.algorithms.Algorithm;
import com.auth0.jwt.interfaces.DecodedJWT;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.authority.SimpleGrantedAuthority;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.net.URL;
import java.security.interfaces.RSAPublicKey;
import java.util.Collections;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.TimeUnit;

public class JwtAuthenticationFilter extends OncePerRequestFilter {

    private static final Logger logger = LoggerFactory.getLogger(JwtAuthenticationFilter.class);
    private static final String ISSUER = "https://okteto.auth0.com/";
    private static final String AUDIENCE = "https://okteto.auth0.com/api/v2/";
    private static final String USERINFO_ENDPOINT = "https://okteto.auth0.com/userinfo";

    private final JwkProvider jwkProvider;
    private final HttpClient httpClient;
    private final ObjectMapper objectMapper;
    private final Map<String, CacheEntry> emailCache;

    private static class CacheEntry {
        final String email;
        final long expiresAt;

        CacheEntry(String email, long ttlMillis) {
            this.email = email;
            this.expiresAt = System.currentTimeMillis() + ttlMillis;
        }

        boolean isExpired() {
            return System.currentTimeMillis() > expiresAt;
        }
    }

    public JwtAuthenticationFilter() throws Exception {
        this.jwkProvider = new JwkProviderBuilder(new URL(ISSUER + ".well-known/jwks.json"))
                .cached(10, 24, TimeUnit.HOURS)
                .rateLimited(10, 1, TimeUnit.MINUTES)
                .build();
        this.httpClient = HttpClient.newBuilder()
                .connectTimeout(java.time.Duration.ofSeconds(5))
                .build();
        this.objectMapper = new ObjectMapper();
        this.emailCache = new ConcurrentHashMap<>();
    }

    private String getUserEmailFromAuth0(String accessToken) {
        // Check cache first
        CacheEntry cached = emailCache.get(accessToken);
        if (cached != null && !cached.isExpired()) {
            return cached.email;
        }

        try {
            HttpRequest request = HttpRequest.newBuilder()
                    .uri(URI.create(USERINFO_ENDPOINT))
                    .header("Authorization", "Bearer " + accessToken)
                    .GET()
                    .build();

            HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());

            if (response.statusCode() == 200) {
                JsonNode userInfo = objectMapper.readTree(response.body());
                String email = userInfo.has("email") ? userInfo.get("email").asText() : null;

                if (email != null && !email.isEmpty()) {
                    // Cache for 5 minutes
                    emailCache.put(accessToken, new CacheEntry(email, 5 * 60 * 1000));
                    return email;
                }
            } else {
                logger.warn("Userinfo endpoint returned status: {}", response.statusCode());
            }
        } catch (Exception e) {
            logger.error("Error fetching userinfo from Auth0: {}", e.getMessage());
        }

        return null;
    }

    @Override
    protected void doFilterInternal(HttpServletRequest request, HttpServletResponse response, FilterChain filterChain)
            throws ServletException, IOException {

        String authHeader = request.getHeader("Authorization");

        if (authHeader == null || !authHeader.startsWith("Bearer ")) {
            logger.warn("Missing or invalid Authorization header");
            response.setStatus(HttpServletResponse.SC_UNAUTHORIZED);
            response.setContentType("application/json");
            response.getWriter().write("{\"message\":\"Missing or invalid Authorization header\"}");
            return;
        }

        String token = authHeader.substring(7);

        try {
            DecodedJWT jwt = JWT.decode(token);

            // Get the public key from JWKS
            Jwk jwk = jwkProvider.get(jwt.getKeyId());
            RSAPublicKey publicKey = (RSAPublicKey) jwk.getPublicKey();

            // Verify the token
            Algorithm algorithm = Algorithm.RSA256(publicKey, null);
            algorithm.verify(jwt);

            // Verify issuer and audience
            if (!ISSUER.equals(jwt.getIssuer())) {
                throw new SecurityException("Invalid issuer");
            }

            if (!jwt.getAudience().contains(AUDIENCE)) {
                throw new SecurityException("Invalid audience");
            }

            // Check expiration
            if (jwt.getExpiresAt().getTime() < System.currentTimeMillis()) {
                throw new SecurityException("Token expired");
            }

            // Try to get email from custom claims in the token
            String email = null;
            String emailClaim = jwt.getClaim("email").asString();
            String namespacedEmailClaim = jwt.getClaim("https://okteto.auth0.com/email").asString();

            if (emailClaim != null && !emailClaim.isEmpty()) {
                email = emailClaim;
            } else if (namespacedEmailClaim != null && !namespacedEmailClaim.isEmpty()) {
                email = namespacedEmailClaim;
            }

            // If email not found in token, fetch from Auth0 userinfo endpoint
            if (email == null || email.isEmpty()) {
                email = getUserEmailFromAuth0(token);
            }

            // Last fallback to subject if no email found
            if (email == null || email.isEmpty()) {
                email = jwt.getSubject();
            }

            // Create authentication object and store in security context
            UsernamePasswordAuthenticationToken authentication =
                new UsernamePasswordAuthenticationToken(
                    email,
                    null,
                    Collections.singletonList(new SimpleGrantedAuthority("ROLE_USER"))
                );

            SecurityContextHolder.getContext().setAuthentication(authentication);

            logger.info("Successfully authenticated user: {}", email);

            filterChain.doFilter(request, response);

        } catch (Exception e) {
            logger.error("JWT validation failed: {}", e.getMessage());
            response.setStatus(HttpServletResponse.SC_UNAUTHORIZED);
            response.setContentType("application/json");
            response.getWriter().write("{\"message\":\"Failed to validate JWT: " + e.getMessage() + "\"}");
        }
    }
}
