package main

import (
	"fmt"
	"gotrack/db"
	"gotrack/tvdbapi"
	"log"
	"net/http"
	"os"
)

var (
	jwtSecret = []byte("super-secret-token")
)

func main() {
	// read environment variables
	TVDB_URL := os.Getenv("TVDB_URL")
	TVDB_TOKEN := os.Getenv("TVDB_TOKEN")
	fmt.Println("TVDB_URL:", TVDB_URL, "TVDB_TOKEN:", TVDB_TOKEN)

	tvdbapi.Setup(TVDB_URL, TVDB_TOKEN)

	SMTP_URL := os.Getenv("SMTP_URL")
	SMTP_PORT := os.Getenv("SMTP_PORT")
	SMTP_USERNAME := os.Getenv("SMTP_USERNAME")
	SMTP_PASSWORD := os.Getenv("SMTP_PASSWORD")
	SMTP_FROM := os.Getenv("SMTP_FROM")
	fmt.Println("SMTP_URL:", SMTP_URL, "SMTP_PORT:", SMTP_PORT, "SMTP_USERNAME:", SMTP_USERNAME, "SMTP_PASSWORD:", SMTP_PASSWORD, "SMTP_FROM:", SMTP_FROM)

	err := SetupMail(SMTP_URL, SMTP_PORT, SMTP_USERNAME, SMTP_PASSWORD, SMTP_FROM)
	if err != nil {
		log.Fatalf("Failed to setup mail: %v", err)
	}

	db.Setup()

	// root - login page, if authed, redirect to other page
	http.HandleFunc("GET /", loggingMiddleware(rootHandler))
	http.HandleFunc("POST /login", loggingMiddleware(loginHandler))
	http.HandleFunc("GET /login", loggingMiddleware(authenticateHandler))

	// home
	http.HandleFunc("GET /home", loggingMiddleware(authMiddleware(homeHandler)))

	// search results
	http.HandleFunc("POST /search", loggingMiddleware(authMiddleware(searchHandler)))

	// add show page
	http.HandleFunc("GET /show/add", loggingMiddleware(authMiddleware(addShowHandler)))
	http.HandleFunc("GET /show/remove", loggingMiddleware(authMiddleware(removeShowHandler)))

	http.HandleFunc("GET /show/details", loggingMiddleware(authMiddleware(showDetailsHandler)))

	http.HandleFunc("GET /show/list", loggingMiddleware(authMiddleware(showListHandler)))

	http.HandleFunc("GET /show/watched", loggingMiddleware(authMiddleware(watchedHandler)))
	http.HandleFunc("GET /show/unwatched", loggingMiddleware(authMiddleware(unwatchedHandler)))

	http.HandleFunc("GET /favicon.png", faviconHandler)

	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type userEmail struct{}

func getUserEmail(r *http.Request) (string, bool) {
	email, ok := r.Context().Value(userEmail{}).(string)
	return email, ok
}

type userID struct{}

func getUserID(r *http.Request) (int64, bool) {
	id, ok := r.Context().Value(userID{}).(int64)
	return id, ok
}
