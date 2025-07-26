package routes

import (
	"log"
	"net/http"
	"strconv"

	"github.com/jccroft1/goshowtrack/auth"
	"github.com/jccroft1/goshowtrack/db"
	"github.com/jccroft1/goshowtrack/tvdbapi"
)

func ShowDetailsHandler(w http.ResponseWriter, r *http.Request) {
	showIDStr := r.URL.Query().Get("id")
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

	showDetails, err := tvdbapi.GetShowDetails(showID, false)
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

	added := userHasAddedShow(userID, showID)

	watchedSeasons := 0
	_ = db.Connection.QueryRow(`SELECT season_number FROM user_seasons WHERE user_id = ? AND show_id = ?`, userID, showID).Scan(&watchedSeasons)

	unwatchedSeasons, _ := hasSomethingToWatch(showDetails.Seasons, watchedSeasons)

	type Season struct {
		Number    int
		Episodes  int
		StartDate string // e.g., "2023-01-15"
		EndDate   string

		Watched  bool
		Released bool
	}
	type ShowData struct {
		ID          int
		Name        string
		AirDate     string
		Description string
		Poster      string
		Status      string // "Continuing" or "Ended"
		Seasons     []Season
		Unwatched   int
	}
	type Data struct {
		Added    bool
		ShowData ShowData
	}

	// Fetch season data from TVDB API
	showData := ShowData{
		ID:          showDetails.ID,
		Name:        showDetails.Name,
		AirDate:     showDetails.AirDate,
		Description: showDetails.Description,
		Poster:      showDetails.PosterPath,
		Status:      showDetails.Status,
		Seasons:     []Season{}, // Initialize with empty slice
		Unwatched:   len(unwatchedSeasons),
	}

	for _, season := range showDetails.Seasons {

		watched := season.Number <= watchedSeasons
		showData.Seasons = append(showData.Seasons, Season{
			Number:    season.Number,
			Episodes:  season.EpisodeCount,
			StartDate: season.AirDate,
			EndDate:   season.LastAirDate,
			Watched:   watched,
			Released:  isReleased(season.LastAirDate),
		})
	}

	data := Data{
		Added:    added,
		ShowData: showData,
	}

	renderTemplate(w, "showDetails", data)
}
