package com.okteto.rent.session;

import jakarta.servlet.http.Cookie;
import jakarta.servlet.http.HttpServletRequest;

import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.Base64;
import java.util.HexFormat;

/**
 * Reads the session cookie issued by the api service. The cookie holds the
 * user email plus an HMAC signature computed with a shared secret.
 */
public final class SessionCookie {
    private static final String COOKIE_NAME = "movies_session";
    private static final String DEFAULT_SECRET = "okteto-movies-demo";

    private SessionCookie() {
    }

    /** Returns the email of the logged in user, or null when there is no valid session. */
    public static String email(HttpServletRequest request) {
        Cookie[] cookies = request.getCookies();
        if (cookies == null) {
            return null;
        }

        for (Cookie cookie : cookies) {
            if (COOKIE_NAME.equals(cookie.getName())) {
                return decode(cookie.getValue());
            }
        }

        return null;
    }

    private static String decode(String value) {
        int separator = value.lastIndexOf('.');
        if (separator < 0) {
            return null;
        }

        String payload = value.substring(0, separator);
        String signature = value.substring(separator + 1);

        String email;
        try {
            email = new String(Base64.getUrlDecoder().decode(payload), StandardCharsets.UTF_8);
        } catch (IllegalArgumentException e) {
            return null;
        }

        if (!MessageDigest.isEqual(sign(email).getBytes(StandardCharsets.UTF_8),
                signature.getBytes(StandardCharsets.UTF_8))) {
            return null;
        }

        return email;
    }

    private static String sign(String value) {
        String secret = System.getenv("SESSION_SECRET");
        if (secret == null || secret.isEmpty()) {
            secret = DEFAULT_SECRET;
        }

        try {
            Mac mac = Mac.getInstance("HmacSHA256");
            mac.init(new SecretKeySpec(secret.getBytes(StandardCharsets.UTF_8), "HmacSHA256"));
            return HexFormat.of().formatHex(mac.doFinal(value.getBytes(StandardCharsets.UTF_8)));
        } catch (Exception e) {
            throw new IllegalStateException("unable to sign the session cookie", e);
        }
    }
}
