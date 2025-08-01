package routes

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/jccroft1/goshowtrack/auth"
	"github.com/jccroft1/goshowtrack/db"
	"github.com/jccroft1/goshowtrack/tvdbapi"
)

// AllHandler lists every show the user has added
func AllHandler(w http.ResponseWriter, req *http.Request) {
	sortType := req.URL.Query().Get("sort")
	if sortType == "" {
		sortType = "name"
	}

	op := func(userID int64, show *tvdbapi.ShowDetail) (bool, ShowData) {
		watchedSeasons := 0
		err := db.Connection.QueryRow(`SELECT season_number FROM user_seasons WHERE user_id = ? AND show_id = ?;`, userID, show.ID).Scan(&watchedSeasons)
		if err != nil && err != sql.ErrNoRows {
			log.Println("Failed to get shows", err)
			return false, ShowData{}
		}

		newShowData := ShowData{
			ID:          show.ID,
			Name:        show.Name,
			AirDate:     show.AirDate,
			Description: show.Description,
			Poster:      show.PosterPath,
			Status:      show.Status,
			SeasonCount: len(show.Seasons),

			Order: show.Name,
		}

		if show.Status == "Returning Series" {
			newShowData.Status = getReturningInfo(*show)
		}

		unwatchedSeasons, _ := hasSomethingToWatch(show.Seasons, watchedSeasons)
		newShowData.Unwatched = len(unwatchedSeasons)

		finished := isFinished(*show)

		switch sortType {
		case "first_release":
			newShowData.Order = show.AirDate
		case "watch_status":
			if watchedSeasons > 0 && len(unwatchedSeasons) > 0 {
				// started, something to watch
				newShowData.Order = "0"
				if finished {
					newShowData.Order += "0"
				} else {
					newShowData.Order += "1"
				}
				newShowData.Order += show.Seasons[watchedSeasons].AirDate
			} else if len(unwatchedSeasons) > 0 {
				// not started
				newShowData.Order = "1"
				if finished {
					newShowData.Order += "0"
				} else {
					newShowData.Order += "1"
				}
				newShowData.Order += show.Seasons[watchedSeasons].AirDate
			} else {
				// watched all
				newShowData.Order = "2"
				if !finished {
					// coming back
					newShowData.Order += "0"
				} else {
					newShowData.Order += "1"
				}
				newShowData.Order += show.AirDate
			}
		}

		return true, newShowData
	}

	listHandler(w, req, op, sortType)
}

// HomeHandler lists unfinished shows the user can watch
func HomeHandler(w http.ResponseWriter, req *http.Request) {
	op := func(userID int64, show *tvdbapi.ShowDetail) (bool, ShowData) {
		watchedSeasons := 0
		err := db.Connection.QueryRow(`SELECT season_number FROM user_seasons WHERE user_id = ? AND show_id = ?`, userID, show.ID).Scan(&watchedSeasons)
		if err != nil && err != sql.ErrNoRows {
			log.Println("Failed to get shows", err)
			return false, ShowData{}
		}

		if watchedSeasons <= 0 {
			return false, ShowData{}
		}

		unwatchedSeasons, somethingToWatch := hasSomethingToWatch(show.Seasons, watchedSeasons)

		if !somethingToWatch {
			return false, ShowData{}
		}

		newShowData := ShowData{
			ID:          show.ID,
			Name:        show.Name,
			AirDate:     show.AirDate,
			Description: show.Description,
			Poster:      show.PosterPath,
			Status:      show.Status,
			SeasonCount: len(show.Seasons),

			Unwatched: len(unwatchedSeasons),
		}
		if show.Status == "Returning Series" {
			newShowData.Status = getReturningInfo(*show)
		}

		finished := isFinished(*show)

		if finished {
			newShowData.Order = "0"
		} else {
			newShowData.Order = "1"
		}
		newShowData.Order += show.Seasons[watchedSeasons].AirDate

		return true, newShowData
	}

	listHandler(w, req, op, "")
}

// StartHandler lists shows the user can start watching
func StartHandler(w http.ResponseWriter, req *http.Request) {
	op := func(userID int64, show *tvdbapi.ShowDetail) (bool, ShowData) {
		watchedSeasons := 0
		err := db.Connection.QueryRow(`SELECT season_number FROM user_seasons WHERE user_id = ? AND show_id = ?`, userID, show.ID).Scan(&watchedSeasons)
		if err != nil && err != sql.ErrNoRows {
			log.Println("Failed to get shows", err)
			return false, ShowData{}
		}
		if watchedSeasons > 0 {
			return false, ShowData{}
		}

		unwatchedSeasons, somethingToWatch := hasSomethingToWatch(show.Seasons, watchedSeasons)

		if !somethingToWatch {
			return false, ShowData{}
		}

		newShowData := ShowData{
			ID:          show.ID,
			Name:        show.Name,
			AirDate:     show.AirDate,
			Description: show.Description,
			Poster:      show.PosterPath,
			Status:      show.Status,
			SeasonCount: len(show.Seasons),

			Unwatched: len(unwatchedSeasons),
		}
		if show.Status == "Returning Series" {
			newShowData.Status = getReturningInfo(*show)
		}

		finished := isFinished(*show)

		if finished {
			newShowData.Order = "0"
		} else {
			newShowData.Order = "1"
		}
		newShowData.Order += show.Seasons[watchedSeasons].AirDate

		return true, newShowData
	}

	listHandler(w, req, op, "")
}

func ComingSoonHandler(w http.ResponseWriter, req *http.Request) {
	op := func(userID int64, show *tvdbapi.ShowDetail) (bool, ShowData) {
		watchedSeasons := 0
		err := db.Connection.QueryRow(`SELECT season_number FROM user_seasons WHERE user_id = ? AND show_id = ?`, userID, show.ID).Scan(&watchedSeasons)
		if err != nil && err != sql.ErrNoRows {
			log.Println("Failed to get shows", err)
			return false, ShowData{}
		}
		_, somethingToWatch := hasSomethingToWatch(show.Seasons, watchedSeasons)

		if somethingToWatch {
			return false, ShowData{}
		}

		finished := isFinished(*show)
		if finished {
			return false, ShowData{}
		}

		newShowData := ShowData{
			ID:          show.ID,
			Name:        show.Name,
			AirDate:     show.AirDate,
			Description: show.Description,
			Poster:      show.PosterPath,
			Status:      show.Status,
			SeasonCount: len(show.Seasons),

			Order: "9999",
		}
		if show.Status == "Returning Series" {
			newShowData.Status = getReturningInfo(*show)
		}

		// get the air data of the next season that's not yet released
		for _, season := range show.Seasons {
			if isReleased(season.LastAirDate) {
				// 0/1 - hack to put the shows with an unreleased season with a known date first
				newShowData.Order = "1" + season.LastAirDate
				continue
			}

			if season.LastAirDate != "" {
				newShowData.Order = "0" + season.LastAirDate
			}

			break
		}

		return true, newShowData
	}

	listHandler(w, req, op, "")
}

func listHandler(w http.ResponseWriter, r *http.Request, op func(int64, *tvdbapi.ShowDetail) (bool, ShowData), sort string) {
	userID, ok := auth.GetUserID(r)
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

	// Fetch all shows for the user
	var showIDs []int
	for showResults.Next() {
		var showID int
		err := showResults.Scan(&showID)
		if err != nil {
			log.Println("Error scanning row: ", err)
			continue
		}
		showIDs = append(showIDs, showID)
	}
	showResults.Close()

	var list []ShowData
	for _, showID := range showIDs {
		show, err := tvdbapi.GetShowDetails(showID, false)
		if err != nil {
			log.Println("Error getting show details: ", err)
			continue
		}

		add, newShow := op(userID, show)
		if !add {
			continue
		}

		list = append(list, newShow)

	}

	type ListData struct {
		Sort string
		List []ShowData
	}

	orderShows(list)

	// Render home page
	renderTemplate(w, "showsList", ListData{Sort: sort, List: list})
}
