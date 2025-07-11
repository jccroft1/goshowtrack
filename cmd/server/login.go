package main

import (
	"fmt"
	"gotrack/db"
	"log"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	gomail "gopkg.in/gomail.v2"
)

var (
	server   string
	port     int
	username string
	password string
	from     string
)

func SetupMail(_server string, _port string, _username string, _password string, _from string) error {
	if server != "" {
		log.Println("No server provided, SMTP disabled. ")
		return nil
	}
	server = _server
	var err error
	port, err = strconv.Atoi(_port)
	if err != nil {
		return fmt.Errorf("invalid port provided: %s", _port)
	}
	username = _username
	password = _password
	from = _from

	return nil
}

func sendMagicLink(email, token string) error {
	link := fmt.Sprintf("http://localhost:8080/login?token=%s", token)

	fmt.Println("Login link", link)

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Your Magic Login Link")
	m.SetBody("text/plain", "Click to sign in: "+link)

	if server == "" {
		return nil
	}

	d := gomail.NewDialer(server, port, username, password)
	return d.DialAndSend(m)
}

func generateAndStoreToken(email string) (string, error) {
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
