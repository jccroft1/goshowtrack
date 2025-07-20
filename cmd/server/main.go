package main

import (
	"fmt"
	"gotrack/auth"
	"gotrack/db"
	"gotrack/mail"
	"gotrack/routes"
	"gotrack/tvdbapi"
	"log"
	"net/http"
	"os"
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

	err := mail.Setup(SMTP_URL, SMTP_PORT, SMTP_USERNAME, SMTP_PASSWORD, SMTP_FROM)
	if err != nil {
		log.Fatalf("Failed to setup mail: %v", err)
	}

	db.Setup()

	http.HandleFunc("GET /favicon.png", routes.FaviconHandler)

	// login
	http.HandleFunc("GET /", loggingMiddleware(routes.RootHandler))
	http.HandleFunc("POST /login", loggingMiddleware(routes.LoginHandler))
	http.HandleFunc("GET /login", loggingMiddleware(routes.AuthenticateHandler))
	http.HandleFunc("GET /logout", loggingMiddleware(routes.LogoutHandler))

	// main pages
	http.HandleFunc("GET /home", loggingMiddleware(auth.Middleware(routes.HomeHandler)))
	http.HandleFunc("GET /search", loggingMiddleware(auth.Middleware(routes.SearchHandler)))
	http.HandleFunc("POST /search", loggingMiddleware(auth.Middleware(routes.SearchResultsHandler)))
	http.HandleFunc("GET /show/list", loggingMiddleware(auth.Middleware(routes.ShowListHandler)))
	http.HandleFunc("GET /show/details", loggingMiddleware(auth.Middleware(routes.ShowDetailsHandler)))

	// show actions
	// TODO: Make these POST requests
	http.HandleFunc("GET /show/add", loggingMiddleware(auth.Middleware(routes.AddShowHandler)))
	http.HandleFunc("GET /show/remove", loggingMiddleware(auth.Middleware(routes.RemoveShowHandler)))
	http.HandleFunc("GET /show/watched", loggingMiddleware(auth.Middleware(routes.WatchedHandler)))
	http.HandleFunc("GET /show/unwatched", loggingMiddleware(auth.Middleware(routes.UnwatchedHandler)))

	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
