package routes

import (
	"log"
	"net/http"
	"strings"

	"github.com/jccroft1/goshowtrack/auth"
	"github.com/jccroft1/goshowtrack/tvdbapi"
)

func BulkAddHandler(w http.ResponseWriter, req *http.Request) {
	userID, ok := auth.GetUserID(req)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	query := req.FormValue("query")
	if query == "" {
		http.Error(w, "Query is required", http.StatusBadRequest)
		return
	}

	// for each line in query, search TVDB and add to user's shows
	lines := strings.Split(query, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		shows, err := tvdbapi.SearchShow(line)
		if err != nil {
			http.Error(w, "Failed to search TVDB", http.StatusInternalServerError)
			return
		}

		// no results found
		if len(shows) == 0 {
			continue
		}

		// add the first result
		showDetails, err := tvdbapi.GetShowDetails(shows[0].ID, false)
		if err != nil {
			log.Println("Error searching TVDB: ", err)
			http.Error(w, "Error searching TVDB", http.StatusInternalServerError)
			return
		}

		addShow(userID, showDetails)
	}

	// redirect to /all page
	http.Redirect(w, req, "/all", http.StatusSeeOther)
}
