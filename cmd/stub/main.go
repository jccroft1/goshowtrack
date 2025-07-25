package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/search/tv", searchHandler)
	http.HandleFunc("/tv/", detailsHandler)

	fmt.Println("Server running on http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	// read the responses/discover.json file
	jsonFile, err := os.ReadFile("responses/search.json")
	if err != nil {
		log.Println(err)
		return
	}

	w.Write(jsonFile)
}

func detailsHandler(w http.ResponseWriter, r *http.Request) {
	// read the responses/discover.json file
	jsonFile, err := os.ReadFile("responses/details.json")
	if err != nil {
		log.Println(err)
		return
	}

	w.Write(jsonFile)
}
