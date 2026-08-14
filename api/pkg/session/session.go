package session

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"strings"
)

const (
	// UserCookie holds the email of the logged in user.
	UserCookie = "movies_session"
	// AdminCookie is set when the admin panel credentials are accepted.
	AdminCookie = "movies_admin"
	adminValue  = "admin"
)

// ErrNoSession is returned when the request carries no valid session.
var ErrNoSession = errors.New("no session")

func secret() []byte {
	if s := os.Getenv("SESSION_SECRET"); s != "" {
		return []byte(s)
	}
	return []byte("okteto-movies-demo")
}

func sign(value string) string {
	mac := hmac.New(sha256.New, secret())
	mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}

// Encode returns a tamper proof cookie value for the given payload.
func Encode(value string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(value)) + "." + sign(value)
}

// Decode validates a cookie value and returns its payload.
func Decode(cookie string) (string, error) {
	parts := strings.SplitN(cookie, ".", 2)
	if len(parts) != 2 {
		return "", ErrNoSession
	}

	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", ErrNoSession
	}

	if !hmac.Equal([]byte(sign(string(raw))), []byte(parts[1])) {
		return "", ErrNoSession
	}

	return string(raw), nil
}

// User returns the email of the user making the request.
func User(r *http.Request) (string, error) {
	cookie, err := r.Cookie(UserCookie)
	if err != nil {
		return "", ErrNoSession
	}
	return Decode(cookie.Value)
}

// IsAdmin reports whether the request carries a valid admin session.
func IsAdmin(r *http.Request) bool {
	cookie, err := r.Cookie(AdminCookie)
	if err != nil {
		return false
	}

	value, err := Decode(cookie.Value)
	return err == nil && value == adminValue
}

// Set writes a session cookie for the given user.
func Set(w http.ResponseWriter, email string) {
	write(w, UserCookie, Encode(email))
}

// SetAdmin writes the admin session cookie.
func SetAdmin(w http.ResponseWriter) {
	write(w, AdminCookie, Encode(adminValue))
}

// Clear removes a session cookie.
func Clear(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{Name: name, Value: "", Path: "/", MaxAge: -1})
}

func write(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7 * 24 * 60 * 60,
	})
}
