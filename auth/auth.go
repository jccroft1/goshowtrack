package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/jccroft1/goshowtrack/db"
)

var (
	disableAuth bool
)

func Setup(_disableAuth bool) {
	disableAuth = _disableAuth
}

func Validate(req *http.Request) (int64, string, bool) {
	if disableAuth {
		return 1, "john.doe@example.com", true
	}

	jwt := req.Header.Get("Cf-Access-Jwt-Assertion")
	if jwt == "" {
		log.Println("No JWT provided")
		return 0, "", false
	}

	// Split the JWT: header.payload.signature
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		log.Println("Invalid JWT format")
		return 0, "", false
	}

	// Decode the payload (2nd part)
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		log.Println("Failed to decode payload")
		return 0, "", false
	}

	// Parse payload JSON
	var payload map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		log.Println("Invalid JSON in payload")
		return 0, "", false
	}

	// Get email
	email, ok := payload["email"].(string)
	if !ok {
		log.Println("Email not found in token")
		return 0, "", false
	}

	// Store user if new
	result, err := db.Connection.Exec(`INSERT OR IGNORE INTO users (email) VALUES (?)`, email)
	if err != nil {
		return 0, "", false
	}
	userID, err := result.LastInsertId()
	if err != nil {
		return 0, "", false
	}
	if userID == 0 {
		// fetch user if existing
		userResults := db.Connection.QueryRow(`SELECT id FROM users WHERE email = ?`, email)
		if userResults.Err() != nil {
			return 0, "", false
		}
		userResults.Scan(&userID)
	}

	return userID, email, true
}

type userEmail struct{}

func GetUserEmail(r *http.Request) (string, bool) {
	email, ok := r.Context().Value(userEmail{}).(string)
	return email, ok
}

type userID struct{}

func GetUserID(r *http.Request) (int64, bool) {
	id, ok := r.Context().Value(userID{}).(int64)
	return id, ok
}

func Middleware(next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, email, ok := Validate(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userEmail{}, email)
		ctx = context.WithValue(ctx, userID{}, id)

		next(w, r.WithContext(ctx))
	})
}
