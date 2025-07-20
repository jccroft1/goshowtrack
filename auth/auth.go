package auth

import (
	"context"
	"fmt"
	"gotrack/db"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
)

var (
	jwtSecret = []byte("super-secret-token")
)

// TODO: Move auth login to package
func Validate(req *http.Request) (int64, string, bool) {
	tokenCookie, err := req.Cookie("token")
	if err != nil {
		return 0, "", false
	}

	tokenStr := tokenCookie.Value
	if tokenStr == "" {
		return 0, "", false
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return 0, "", false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, "", false
	}

	// TODO: Improve error handling here
	email := claims["email"].(string)
	if email == "" {
		return 0, "", false
	}

	userID := int64(claims["id"].(float64))
	if userID == 0 {
		return 0, "", false
	}

	return userID, email, true
}

func GenerateAndStoreToken(email string) (string, error) {
	// Store user if new
	result, err := db.Connection.Exec(`INSERT OR IGNORE INTO users (email) VALUES (?)`, email)
	if err != nil {
		return "", err
	}
	userID, err := result.LastInsertId()
	if err != nil {
		return "", err
	}
	if userID == 0 {
		userResults := db.Connection.QueryRow(`SELECT id FROM users WHERE email = ?`, email)
		if userResults.Err() != nil {
			return "", fmt.Errorf("error retrieving user ID: %v", userResults.Err())
		}
		userResults.Scan(&userID)
	}
	fmt.Println("User ID:", userID)

	// Generate token
	tokenStr, err := createSessionJWT(userID, email)
	if err != nil {
		return "", err
	}

	// Store in DB
	expiresAt := time.Now().Add(15 * time.Minute).Unix()
	_, err = db.Connection.Exec(`
        INSERT INTO magic_links (token, email, expires_at) VALUES (?, ?, ?)
    `, tokenStr, email, expiresAt)
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func createSessionJWT(id int64, email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"id":    id,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString(jwtSecret)
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
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		ctx := context.WithValue(r.Context(), userEmail{}, email)
		ctx = context.WithValue(ctx, userID{}, id)

		next(w, r.WithContext(ctx))
	})
}
