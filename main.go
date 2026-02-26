package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	dbClose := db.Setup()
	defer dbClose()

	TVDB_TOKEN := os.Getenv("TVDB_TOKEN")
	tvdbapi.Setup(TVDB_TOKEN)

	DISABLE_AUTH := os.Getenv("DISABLE_AUTH")
	auth.Setup(DISABLE_AUTH == "true")

	// Setup server
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("./assets/"))
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", fs))

	// main pages
	mux.HandleFunc("GET /", logging.Middleware(auth.Middleware(routes.HomeHandler)))
	mux.HandleFunc("GET /about", logging.Middleware(auth.Middleware(routes.AboutHandler)))
	mux.HandleFunc("GET /start", logging.Middleware(auth.Middleware(routes.StartHandler)))
	mux.HandleFunc("GET /comingsoon", logging.Middleware(auth.Middleware(routes.ComingSoonHandler)))
	mux.HandleFunc("GET /all", logging.Middleware(auth.Middleware(routes.AllHandler)))

	// search pages
	mux.HandleFunc("GET /search", logging.Middleware(auth.Middleware(routes.SearchHandler)))
	mux.HandleFunc("POST /search", logging.Middleware(auth.Middleware(routes.SearchResultsHandler)))
	mux.HandleFunc("POST /bulk_add", logging.Middleware(auth.Middleware(routes.BulkAddHandler)))
	mux.HandleFunc("GET /show/details", logging.Middleware(auth.Middleware(routes.ShowDetailsHandler)))
	mux.HandleFunc("GET /autofill", logging.Middleware(auth.Middleware(routes.AutofillHandler)))

	// show actions
	mux.HandleFunc("GET /show/add", logging.Middleware(auth.Middleware(routes.AddShowHandler)))
	mux.HandleFunc("GET /show/remove", logging.Middleware(auth.Middleware(routes.RemoveShowHandler)))
	mux.HandleFunc("GET /show/watched", logging.Middleware(auth.Middleware(routes.WatchedHandler)))
	mux.HandleFunc("GET /show/unwatched", logging.Middleware(auth.Middleware(routes.UnwatchedHandler)))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Run server
	go func() {
		log.Println("Server running on http://localhost:8080")

		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Wait for signal
	<-stop
	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown HTTP server
	err := server.Shutdown(ctx)
	if err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server shutdown complete.")
}
