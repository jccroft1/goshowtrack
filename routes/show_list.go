package routes

import (
	"gotrack/auth"
	"gotrack/db"
	"gotrack/tvdbapi"
	"log"
	"net/http"
)

// TODO: Make response writer and request parameters use consist naming conventions for all handlers
func ShowListHandler(w http.ResponseWriter, req *http.Request) {
	userID, ok := auth.GetUserID(req)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// SQL to fetch the User's Shows
	showResults, err := db.Connection.Query(`SELECT show_id FROM user_shows WHERE user_id = ?`, userID)
	if err != nil {
		log.Println("Error fetch user show list: ", err)
		http.Error(w, "Failed to fetch user shows", http.StatusInternalServerError)
		return
	}
	defer showResults.Close()

	var shows []ShowData
	for showResults.Next() {
		var showID int
		err := showResults.Scan(&showID)
		if err != nil {
			log.Println("Error scanning row: ", err)
			continue
		}

		// tvdb API call to get the TV Show details
		show, err := tvdbapi.GetShowDetails(showID)
		if err != nil {
			log.Println("Error getting show details: ", err)
			continue
		}

		watchedSeason := 0
		_ = db.Connection.QueryRow(`SELECT season_number FROM user_seasons WHERE user_id = ? AND show_id = ?`, userID, showID).Scan(&watchedSeason)
		watchAction, watchActionColor := generateActionText(show.Seasons, watchedSeason)

		newShow := ShowData{
			ID:          show.ID,
			Name:        show.Name,
			AirDate:     show.AirDate,
			Description: show.Description,
			Poster:      show.PosterPath,
			Status:      show.Status,
			SeasonCount: len(show.Seasons),

			Order:            show.Name,
			WatchAction:      watchAction,
			WatchActionColor: watchActionColor,
		}

		if show.Status == "Returning Series" {
			newShow.Status = getReturningInfo(*show)
		}

		shows = append(shows, newShow)
	}

	type SearchData struct {
		List []ShowData
	}
	data := SearchData{
		List: shows,
	}

	orderShows(shows)

	renderTemplate(w, "showsList", data)
}
