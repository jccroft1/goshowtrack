package routes

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/jccroft1/goshowtrack/auth"
	"github.com/jccroft1/goshowtrack/db"
)

func WatchedHandler(w http.ResponseWriter, r *http.Request) {
	userWatchedUpdate(w, r, true)
}

func UnwatchedHandler(w http.ResponseWriter, r *http.Request) {
	userWatchedUpdate(w, r, false)
}

func userWatchedUpdate(w http.ResponseWriter, r *http.Request, watched bool) {
	showIDStr := r.URL.Query().Get("show_id")
	if showIDStr == "" {
		http.Error(w, "No ID provided", http.StatusBadRequest)
		return
	}

	showID, err := strconv.Atoi(showIDStr)
	if err != nil {
		log.Println("Invalid ID provided", err)
		http.Error(w, "Invalid ID provided", http.StatusBadRequest)
		return
	}

	seasonNumberStr := r.URL.Query().Get("season")
	if seasonNumberStr == "" {
		http.Error(w, "No ID provided", http.StatusBadRequest)
		return
	}

	seasonNumber, err := strconv.Atoi(seasonNumberStr)
	if err != nil {
		log.Println("Invalid ID provided", err)
		http.Error(w, "Invalid ID provided", http.StatusBadRequest)
		return
	}

	userID, ok := auth.GetUserID(r)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	_, err = db.Connection.Exec("DELETE FROM user_seasons WHERE user_id = ? AND show_id = ?", userID, showID)
	if err != nil {
		log.Println("Error removing season from user", err)
		http.Error(w, "Error removing season from user", http.StatusInternalServerError)
		return
	}
	if !watched {
		seasonNumber--
	}

	if seasonNumber >= 1 {
		_, err = db.Connection.Exec("INSERT OR IGNORE INTO user_seasons (user_id, show_id, season_number) VALUES (?, ?, ?)", userID, showID, seasonNumber)
		if err != nil {
			log.Println("Error adding season to user", err)
			http.Error(w, "Error adding season to user", http.StatusInternalServerError)
			return
		}
	}

	// redirect to show details page
	http.Redirect(w, r, fmt.Sprintf("/show/details?id=%v", showID), http.StatusSeeOther)
}
