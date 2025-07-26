package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/jccroft1/goshowtrack/auth"
	"github.com/jccroft1/goshowtrack/db"
	"github.com/jccroft1/goshowtrack/logging"
	"github.com/jccroft1/goshowtrack/routes"
	"github.com/jccroft1/goshowtrack/tvdbapi"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime)
	flag.BoolVar(&logging.Verbose, "v", false, "enable verbose logging")
	flag.Parse()

	db.Setup()

	TVDB_TOKEN := os.Getenv("TVDB_TOKEN")
	tvdbapi.Setup(TVDB_TOKEN)

	DISABLE_AUTH := os.Getenv("DISABLE_AUTH")
	auth.Setup(DISABLE_AUTH == "true")

	fs := http.FileServer(http.Dir("./assets/"))
	http.Handle("GET /assets/", http.StripPrefix("/assets/", fs))

	// main pages
	http.HandleFunc("GET /", logging.Middleware(auth.Middleware(routes.HomeHandler)))
	http.HandleFunc("GET /about", logging.Middleware(auth.Middleware(routes.AboutHandler)))
	http.HandleFunc("GET /start", logging.Middleware(auth.Middleware(routes.StartHandler)))
	http.HandleFunc("GET /comingsoon", logging.Middleware(auth.Middleware(routes.ComingSoonHandler)))
	http.HandleFunc("GET /all", logging.Middleware(auth.Middleware(routes.AllHandler)))

	http.HandleFunc("GET /search", logging.Middleware(auth.Middleware(routes.SearchHandler)))
	http.HandleFunc("POST /search", logging.Middleware(auth.Middleware(routes.SearchResultsHandler)))
	http.HandleFunc("POST /bulk_add", logging.Middleware(auth.Middleware(routes.BulkAddHandler)))
	http.HandleFunc("GET /show/details", logging.Middleware(auth.Middleware(routes.ShowDetailsHandler)))

	// show actions
	http.HandleFunc("GET /show/add", logging.Middleware(auth.Middleware(routes.AddShowHandler)))
	http.HandleFunc("GET /show/remove", logging.Middleware(auth.Middleware(routes.RemoveShowHandler)))
	http.HandleFunc("GET /show/watched", logging.Middleware(auth.Middleware(routes.WatchedHandler)))
	http.HandleFunc("GET /show/unwatched", logging.Middleware(auth.Middleware(routes.UnwatchedHandler)))

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
