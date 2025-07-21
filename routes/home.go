package routes

import (
	"fmt"
	"gotrack/auth"
	"gotrack/db"
	"gotrack/tvdbapi"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func OldHomeHandler(w http.ResponseWriter, req *http.Request) {

	userID, ok := auth.GetUserID(req)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	fmt.Println(userID)

	// SQL to fetch the User's Shows
	showResults, err := db.Connection.Query(`SELECT show_id FROM user_shows WHERE user_id = ?`, userID)
	if err != nil {
		log.Println("Error fetch user show list: ", err)
		http.Error(w, "Failed to fetch user shows", http.StatusInternalServerError)
		return
	}
	defer showResults.Close()

	var toWatchContinuing, toWatchFinished, toStartContinuing, toStartFinished, toWaitList []ShowData

	for showResults.Next() {
		var showID int
		err := showResults.Scan(&showID)
		if err != nil {
			log.Println("Error scanning row: ", err)
			continue
		}

		show, err := tvdbapi.GetShowDetails(showID)
		if err != nil {
			log.Println("Error getting show details: ", err)
			continue
		}

		watchedSeasons := 0
		_ = db.Connection.QueryRow(`SELECT season_number FROM user_seasons WHERE user_id = ? AND show_id = ?`, userID, showID).Scan(&watchedSeasons)

		// watchAction, watchActionColor := generateActionText(show.Seasons, watchedSeasons)

		newShowData := ShowData{
			ID:          show.ID,
			Name:        show.Name,
			AirDate:     show.AirDate,
			Description: show.Description,
			Poster:      show.PosterPath,
			Status:      show.Status,
			SeasonCount: len(show.Seasons),

			// WatchAction:      watchAction,
			// WatchActionColor: watchActionColor,
		}
		if show.Status == "Returning Series" {
			newShowData.Status = getReturningInfo(*show)
		}

		finished := isFinished(*show)
		_, somethingToWatch := hasSomethingToWatch(show.Seasons, watchedSeasons)

		if somethingToWatch {
			newShowData.Order = show.Seasons[watchedSeasons].AirDate

			if watchedSeasons > 0 {
				// Ready to watch
				// newShowData.WatchAction = fmt.Sprintf("You've got %v ready to watch", toSeasonComma(toWatchSeasons))
				newShowData.WatchAction, _ = generateActionText(show.Seasons, watchedSeasons)
				newShowData.WatchActionColor = "green"

				if finished {
					toWatchFinished = append(toWatchFinished, newShowData)
				} else {
					toWatchContinuing = append(toWatchContinuing, newShowData)
				}
			} else {
				// Start something new?
				if finished {
					toStartFinished = append(toStartFinished, newShowData)
				} else {
					toStartContinuing = append(toStartContinuing, newShowData)
				}
			}
		} else {
			if !finished {
				toWaitList = append(toWaitList, newShowData)
			}
		}
	}

	type HomeData struct {
		ToWatch []ShowData
		ToStart []ShowData
		ToWait  []ShowData
	}

	orderShows(toWatchFinished)
	orderShows(toWatchContinuing)
	orderShows(toStartFinished)
	orderShows(toStartContinuing)
	orderShows(toWaitList)

	toWatchList := append(toWatchFinished, toWatchContinuing...)
	toStartList := append(toStartFinished, toStartContinuing...)

	// Render home page
	renderTemplate(w, "home", HomeData{ToWatch: toWatchList, ToStart: toStartList, ToWait: toWaitList})
}

func toSeasonComma(list []int) string {
	if len(list) == 0 {
		return ""
	}

	if len(list) == 1 {
		return "season " + strconv.Itoa(list[0])
	}

	if len(list) == 2 {
		return "seasons " + strconv.Itoa(list[0]) + " and " + strconv.Itoa(list[1])
	}

	out := "seasons"
	for _, num := range list[:len(list)-1] {
		out += ", " + strconv.Itoa(num)
	}
	out += ", and " + strconv.Itoa(list[len(list)-1])
	return out
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
	case "ended":
		return true
	case "canceled":
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
