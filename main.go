package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jccroft1/goshowtrack/auth"
	"github.com/jccroft1/goshowtrack/db"
	"github.com/jccroft1/goshowtrack/logging"
	"github.com/jccroft1/goshowtrack/mail"
	"github.com/jccroft1/goshowtrack/routes"
	"github.com/jccroft1/goshowtrack/tvdbapi"
)

func main() {
	// read environment variables

	// TVDB API
	TVDB_URL := os.Getenv("TVDB_URL")
	TVDB_TOKEN := os.Getenv("TVDB_TOKEN")
	fmt.Println("TVDB_URL:", TVDB_URL, "TVDB_TOKEN:", TVDB_TOKEN)

	tvdbapi.Setup(TVDB_URL, TVDB_TOKEN)

	// Mail Server
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
	http.HandleFunc("GET /", logging.Middleware(routes.RootHandler))
	http.HandleFunc("POST /login", logging.Middleware(routes.LoginHandler))
	http.HandleFunc("GET /login", logging.Middleware(routes.AuthenticateHandler))
	http.HandleFunc("GET /logout", logging.Middleware(routes.LogoutHandler))

	// main pages
	http.HandleFunc("GET /home", logging.Middleware(auth.Middleware(routes.HomeHandler)))
	http.HandleFunc("GET /start", logging.Middleware(auth.Middleware(routes.StartHandler)))
	http.HandleFunc("GET /comingsoon", logging.Middleware(auth.Middleware(routes.ComingSoonHandler)))
	http.HandleFunc("GET /all", logging.Middleware(auth.Middleware(routes.AllHandler)))

	http.HandleFunc("GET /search", logging.Middleware(auth.Middleware(routes.SearchHandler)))
	http.HandleFunc("POST /search", logging.Middleware(auth.Middleware(routes.SearchResultsHandler)))
	http.HandleFunc("GET /show/details", logging.Middleware(auth.Middleware(routes.ShowDetailsHandler)))

	// show actions
	// TODO: Make these POST requests
	http.HandleFunc("GET /show/add", logging.Middleware(auth.Middleware(routes.AddShowHandler)))
	http.HandleFunc("GET /show/remove", logging.Middleware(auth.Middleware(routes.RemoveShowHandler)))
	http.HandleFunc("GET /show/watched", logging.Middleware(auth.Middleware(routes.WatchedHandler)))
	http.HandleFunc("GET /show/unwatched", logging.Middleware(auth.Middleware(routes.UnwatchedHandler)))

	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
