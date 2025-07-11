package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// http server stub of an API
	// curl --request GET \
	// --url 'https://api.themoviedb.org/3/discover/tv?include_adult=false&include_null_first_air_dates=false&language=en-US&page=1&sort_by=popularity.desc' \
	// --header 'accept: application/json'
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
