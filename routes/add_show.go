package routes

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/jccroft1/goshowtrack/auth"
	"github.com/jccroft1/goshowtrack/db"
	"github.com/jccroft1/goshowtrack/tvdbapi"
)

func AddShowHandler(w http.ResponseWriter, r *http.Request) {
	userShowUpdate(w, r, true)
}

func RemoveShowHandler(w http.ResponseWriter, r *http.Request) {
	userShowUpdate(w, r, false)
}

func userShowUpdate(w http.ResponseWriter, r *http.Request, add bool) {
	queryStr := r.URL.Query().Get("id")
	if queryStr == "" {
		log.Println("No ID provided")
		http.Error(w, "No ID provided", http.StatusBadRequest)
		return
	}

	query, err := strconv.Atoi(queryStr)
	if err != nil {
		log.Println("Invalid ID provided", err)
		http.Error(w, "Invalid ID provided", http.StatusBadRequest)
		return
	}

	// not strictly necessary but checks the show is valid and loads into cache
	showDetails, err := tvdbapi.GetShowDetails(query, false)
	if err != nil {
		log.Println("Error searching TVDB: ", err)
		http.Error(w, "Error searching TVDB", http.StatusInternalServerError)
		return
	}

	userID, ok := auth.GetUserID(r)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	if add {
		// SQL to add show to user
		err := addShow(userID, showDetails)
		if err != nil {
			log.Println("Error adding show to user:", err)
			http.Error(w, "Failed to add show to user", http.StatusInternalServerError)
			return
		}
	} else {
		// SQL to remove show from user
		sqlQuery := `DELETE FROM user_shows WHERE user_id = ? AND show_id = ?`
		_, err = db.Connection.Exec(sqlQuery, userID, showDetails.ID)
		if err != nil {
			log.Println("Error adding show to user:", err)
			http.Error(w, "Failed to add show to user", http.StatusInternalServerError)
			return
		}
	}

	// redirect to show details page
	http.Redirect(w, r, fmt.Sprintf("/show/details?id=%v", showDetails.ID), http.StatusSeeOther)
}

func addShow(userID int64, show *tvdbapi.ShowDetail) error {
	alreadyAdded := userHasAddedShow(userID, show.ID)
	if alreadyAdded {
		return nil
	}

	sqlQuery := `INSERT INTO user_shows (user_id, show_id) VALUES (?, ?)`

	_, err := db.Connection.Exec(sqlQuery, userID, show.ID)
	if err != nil {
		return fmt.Errorf("error adding show to user: %v", err)
	}

	return nil
}
