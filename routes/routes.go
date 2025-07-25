package routes

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/jccroft1/goshowtrack/db"
	"github.com/jccroft1/goshowtrack/tvdbapi"
)

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	tmplBase := template.New("layout").Funcs(template.FuncMap{
		"dateToYear": dateToYear,
	})

	tmpls := template.Must(tmplBase.ParseFiles(
		"templates/layout.html",
		"templates/"+tmpl+".html",
		"templates/partials/searchBar.html",
		"templates/partials/navBar.html",
	))
	err := tmpls.ExecuteTemplate(w, "layout", data)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

func userHasAddedShow(userID int64, showID int) bool {
	var id int
	err := db.Connection.QueryRow("SELECT user_id FROM user_shows WHERE show_id = ? AND user_id = ?", showID, userID).Scan(&id)

	return err != sql.ErrNoRows
}

func generateActionText(seasons []tvdbapi.Season, watchedSeasons int) (string, string) {
	totalSeasons := getTotalReleasedSeasons(seasons)

	if totalSeasons == 0 {
		return "", ""
		return "This show doesn't have any episodes yet. Stay tuned!", "red"
	}

	if watchedSeasons == totalSeasons {
		return "", ""
		return "You've watched all the available seasons of this show.", "red"
	}

	if watchedSeasons < totalSeasons {
		if totalSeasons-watchedSeasons == 1 {
			return "You have one more season left to watch!", "green"
		}

		return fmt.Sprintf("You've got %v more seasons ready to watch!", totalSeasons-watchedSeasons), "green"
	}

	return "Unknown", "grey"
}

// Render the results to the user
type ShowData struct {
	// TVDB Data
	ID          int
	Name        string
	AirDate     string
	Description string
	Poster      string
	Status      string
	SeasonCount int

	// UI features
	Order            string
	WatchAction      string
	WatchActionColor string
}

func orderShows(shows []ShowData) []ShowData {
	sort.Slice(shows, func(i, j int) bool {
		return shows[i].Order < shows[j].Order
	})
	return shows
}

func dateToYear(s string) string {
	out := ""
	if len(s) > 4 {
		out = s[0:4]
	}
	return out
}

func getTotalReleasedSeasons(seasons []tvdbapi.Season) int {
	releasedSeasons := 0
	for _, season := range seasons {
		if !isReleased(season.LastAirDate) {
			continue
		}

		releasedSeasons++
	}
	return releasedSeasons
}

func isReleased(seasonAirDate string) bool {
	if strings.TrimSpace(seasonAirDate) == "" {
		return false
	}

	parsedAirDate, err := time.Parse("2006-01-02", seasonAirDate)
	if err != nil {
		log.Println("Error parsing finish date:", err)
		return false
	}

	now := time.Now()

	return now.After(parsedAirDate)
}

func hasSomethingToWatch(seasons []tvdbapi.Season, watchedSeasons int) ([]int, bool) {
	if len(seasons) < watchedSeasons {
		return []int{}, false
	}

	toWatch := []int{}
	for _, season := range seasons[watchedSeasons:] {
		if !isReleased(season.LastAirDate) {
			break
		}

		toWatch = append(toWatch, season.Number)
	}

	return toWatch, len(toWatch) > 0
}

func isFinished(show tvdbapi.ShowDetail) bool {
	status := strings.ToLower(show.Status)
	switch status {
	case "returning series":
		return false
	case "ended", "canceled":
		return true
	default:
		return false
	}
}

func getReturningInfo(show tvdbapi.ShowDetail) string {
	for _, season := range show.Seasons {
		if isReleased(season.LastAirDate) {
			continue
		}

		lastAirDate, err := time.Parse("2006-01-02", season.LastAirDate)
		if err != nil {
			break
		}

		countdown := int(time.Until(lastAirDate).Hours() / 24)
		return fmt.Sprintf("Returning in %d days", countdown)
	}

	return "Coming back at some point..."
}
