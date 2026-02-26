package routes

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jccroft1/goshowtrack/auth"
	"github.com/jccroft1/goshowtrack/tvdbapi"
)

func SearchHandler(w http.ResponseWriter, req *http.Request) {
	if req.URL.Query().Get("bulk") == "true" {
		renderTemplate(w, "searchBulk", nil)
		return
	}

	renderTemplate(w, "search", nil)
}

func SearchResultsHandler(w http.ResponseWriter, req *http.Request) {
	userID, ok := auth.GetUserID(req)
	if !ok {
		log.Println("user not authenticated")
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	query := req.FormValue("query")
	if query == "" {
		log.Println("search request missing 'query'")
		http.Error(w, "Query is required", http.StatusBadRequest)
		return
	}

	searchResults, err := tvdbapi.SearchShow(query)
	if err != nil {
		log.Println("search failed", err)
		http.Error(w, "Failed to search TVDB", http.StatusInternalServerError)
		return
	}

	type ShowData struct {
		ID          int
		Name        string
		AirDate     string
		Description string
		Poster      string

		Added bool
	}
	type SearchData struct {
		Query   string
		Results []ShowData
	}

	data := SearchData{
		Results: make([]ShowData, len(searchResults)),
		Query:   query,
	}
	for i, show := range searchResults {

		data.Results[i] = ShowData{
			ID:          show.ID,
			Name:        show.Name,
			AirDate:     show.AirDate,
			Description: show.Description,
			Poster:      show.PosterPath,

			Added: userHasAddedShow(userID, show.ID),
		}

	}

	renderTemplate(w, "searchResults", data)
}

func AutofillHandler(w http.ResponseWriter, r *http.Request) {
	searchName := r.URL.Query().Get("query")

	if len(searchName) < 3 {
		err := json.NewEncoder(w).Encode([]string{})
		if err != nil {
			log.Println("empty autofill handler response failed", err)
			return
		}
		return
	}

	found := tvdbapi.FindPopularShows(searchName)

	err := json.NewEncoder(w).Encode(found)
	if err != nil {
		log.Println("autofill handler response failed", err)
		return
	}
}
