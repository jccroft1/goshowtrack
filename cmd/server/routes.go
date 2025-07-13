package main

import (
	"database/sql"
	"fmt"
	"gotrack/db"
	"gotrack/tvdbapi"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func rootHandler(w http.ResponseWriter, req *http.Request) {
	_, _, ok := validateAuth(req)
	if ok {
		http.Redirect(w, req, "/home", http.StatusFound)
		return
	}

	renderTemplate(w, "login", nil)
}

func authenticateHandler(w http.ResponseWriter, req *http.Request) {
	// read token from query string
	token := req.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	var email string
	var expiresAt int64
	var used int
	err := db.Connection.QueryRow(`
        SELECT email, expires_at, used FROM magic_links WHERE token = ?
    `, token).Scan(&email, &expiresAt, &used)

	if err == sql.ErrNoRows {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	if used != 0 {
		http.Error(w, "Token already used", http.StatusUnauthorized)
		return
	}

	if time.Now().Unix() > expiresAt {
		http.Error(w, "Token expired", http.StatusUnauthorized)
		return
	}

	// Mark token as used
	_, _ = db.Connection.Exec(`UPDATE magic_links SET used = 1 WHERE token = ?`, token)

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   token,
		Path:    "/",
		Expires: time.Now().Add(24 * time.Hour),
	})

	// redirect to home page
	http.Redirect(w, req, "/home", http.StatusFound)
}

func homeHandler(w http.ResponseWriter, req *http.Request) {
	email, ok := getUserEmail(req)
	if !ok {
		http.Error(w, "Invalid user", http.StatusUnauthorized)
		return
	}

	type HelloData struct {
		Email string
		Query string
	}

	// Render home page
	renderTemplate(w, "home", HelloData{Email: email})
}

func loginHandler(w http.ResponseWriter, req *http.Request) {
	email := req.FormValue("email")
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	token, err := generateAndStoreToken(email)
	if err != nil {
		log.Println(err)
		http.Error(w, "Could not create token", http.StatusInternalServerError)
		return
	}

	err = sendMagicLink(email, token)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	// Render the login template or redirect to login page
	renderTemplate(w, "loginRedirect", nil)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	tmpls := template.Must(template.ParseFiles(
		"../../templates/layout.html",
		"../../templates/"+tmpl+".html",
		"../../templates/partials/searchBar.html",
		"../../templates/partials/navBar.html",
	))
	err := tmpls.Execute(w, data)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

// TODO: Move auth login to package
func validateAuth(req *http.Request) (int64, string, bool) {
	tokenCookie, err := req.Cookie("token")
	if err != nil {
		return 0, "", false
	}

	tokenStr := tokenCookie.Value
	if tokenStr == "" {
		return 0, "", false
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return 0, "", false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, "", false
	}

	// TODO: Improve error handling here
	email := claims["email"].(string)
	if email == "" {
		return 0, "", false
	}

	userID := int64(claims["id"].(float64))
	if userID == 0 {
		return 0, "", false
	}

	return userID, email, true
}

func searchHandler(w http.ResponseWriter, req *http.Request) {
	userID, ok := getUserID(req)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	query := req.FormValue("query")
	if query == "" {
		http.Error(w, "Query is required", http.StatusBadRequest)
		return
	}

	searchResults, err := tvdbapi.SearchShow(query)
	if err != nil {
		http.Error(w, "Failed to search TVDB", http.StatusInternalServerError)
		return
	}

	type ShowData struct {
		ID          int
		Name        string
		Year        string
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
		year := ""
		if len(show.AirDate) > 4 {
			year = show.AirDate[0:4]
		}

		data.Results[i] = ShowData{
			ID:          show.ID,
			Name:        show.Name,
			Year:        year,
			Description: show.Description,
			Poster:      show.PosterPath,

			Added: userHasAddedShow(userID, show.ID),
		}

	}

	renderTemplate(w, "searchResults", data)
}

// TODO: Make response writer and request parameters use consist naming conventions for all handlers
func showListHandler(w http.ResponseWriter, req *http.Request) {
	userID, ok := getUserID(req)
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

	// Render the results to the user
	type ShowData struct {
		ID               int
		Name             string
		Year             string
		Description      string
		Poster           string
		Status           string
		SeasonCount      int
		WatchAction      string
		WatchActionColor string
	}
	type SearchData struct {
		List []ShowData
	}

	data := SearchData{
		List: make([]ShowData, 0),
	}
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

		year := ""
		if len(show.AirDate) > 4 {
			year = show.AirDate[0:4]
		}

		seasonCount := 0
		for _, season := range show.Seasons {
			if season.Number == 0 {
				continue
			}
			seasonCount++
		}

		watchedSeason := 0
		_ = db.Connection.QueryRow(`SELECT season_number FROM user_seasons WHERE user_id = ? AND show_id = ?`, userID, showID).Scan(&watchedSeason)
		watchAction, watchActionColor := generateActionText(seasonCount, watchedSeason)

		data.List = append(data.List, ShowData{
			ID:               show.ID,
			Name:             show.Name,
			Year:             year,
			Description:      show.Description,
			Poster:           show.PosterPath,
			Status:           show.Status,
			SeasonCount:      seasonCount,
			WatchAction:      watchAction,
			WatchActionColor: watchActionColor,
		})
	}

	sort.Slice(data.List, func(i, j int) bool {
		return data.List[i].Name < data.List[j].Name
	})

	renderTemplate(w, "showsList", data)
}

func addShowHandler(w http.ResponseWriter, r *http.Request) {
	userShowUpdate(w, r, true)
}

func removeShowHandler(w http.ResponseWriter, r *http.Request) {
	userShowUpdate(w, r, false)
}

func userShowUpdate(w http.ResponseWriter, r *http.Request, add bool) {
	queryStr := r.URL.Query().Get("id")
	if queryStr == "" {
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
	showDetails, err := tvdbapi.GetShowDetails(query)
	if err != nil {
		log.Println("Error searching TVDB: ", err)
		http.Error(w, "Error searching TVDB", http.StatusInternalServerError)
		return
	}

	userID, ok := getUserID(r)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	if add {
		// SQL to add show to user
		alreadyAdded := userHasAddedShow(userID, query)

		if !alreadyAdded {
			sqlQuery := `INSERT INTO user_shows (user_id, show_id) VALUES (?, ?)`

			_, err = db.Connection.Exec(sqlQuery, userID, showDetails.ID)
			if err != nil {
				log.Println("Error adding show to user:", err)
				http.Error(w, "Failed to add show to user", http.StatusInternalServerError)
				return
			}
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

func watchedHandler(w http.ResponseWriter, r *http.Request) {
	userWatchedUpdate(w, r, true)
}

func unwatchedHandler(w http.ResponseWriter, r *http.Request) {
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

	userID, ok := getUserID(r)
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

func showDetailsHandler(w http.ResponseWriter, r *http.Request) {
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

	showDetails, err := tvdbapi.GetShowDetails(showID)
	if err != nil {
		log.Println("Error searching TVDB: ", err)
		http.Error(w, "Error searching TVDB", http.StatusInternalServerError)
		return
	}

	userID, ok := getUserID(r)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	added := userHasAddedShow(userID, showID)

	watchedSeason := 0
	_ = db.Connection.QueryRow(`SELECT season_number FROM user_seasons WHERE user_id = ? AND show_id = ?`, userID, showID).Scan(&watchedSeason)

	type Season struct {
		Number    int
		Episodes  int
		StartDate string // e.g., "2023-01-15"

		Watched bool
	}
	type ShowData struct {
		ID               int
		Name             string
		Year             string
		Description      string
		Poster           string
		Status           string // "Continuing" or "Ended"
		Seasons          []Season
		WatchAction      string
		WatchActionColor string
	}
	type Data struct {
		Added    bool
		ShowData ShowData
	}

	year := ""
	if len(showDetails.AirDate) > 4 {
		year = showDetails.AirDate[0:4]
	}
	// Fetch season data from TVDB API
	showData := ShowData{
		ID:          showDetails.ID,
		Name:        showDetails.Name,
		Year:        year,
		Description: showDetails.Description,
		Poster:      showDetails.PosterPath,
		Status:      showDetails.Status,
		Seasons:     []Season{}, // Initialize with empty slice

	}
	seasonCount := 0
	for _, season := range showDetails.Seasons {
		if season.Number == 0 {
			continue
		}
		seasonCount++
		watched := season.Number <= watchedSeason
		showData.Seasons = append(showData.Seasons, Season{
			Number:    season.Number,
			Episodes:  season.EpisodeCount,
			StartDate: season.AirDate,
			Watched:   watched,
		})
	}
	showData.WatchAction, showData.WatchActionColor = generateActionText(seasonCount, watchedSeason)

	data := Data{
		Added:    added,
		ShowData: showData,
	}

	renderTemplate(w, "showDetails", data)
}

func userHasAddedShow(userID int64, showID int) bool {
	var id int
	err := db.Connection.QueryRow("SELECT user_id FROM user_shows WHERE show_id = ? AND user_id = ?", showID, userID).Scan(&id)

	return err != sql.ErrNoRows
}

func generateActionText(totalSeasons, watchedSeasons int) (string, string) {
	if watchedSeasons == 0 && totalSeasons > 0 {
		return fmt.Sprintf("Great news! You've got all %v seasons ready to watch.", totalSeasons), "green"
	}

	if totalSeasons == 0 {
		return "This show doesn't have any episodes yet. Stay tuned!", "red"
	}

	if watchedSeasons == totalSeasons {
		return "You've already watched all the seasons of this show.", "red"
	}

	if watchedSeasons < totalSeasons {
		if totalSeasons-watchedSeasons == 1 {
			return "You have one more season left to watch!", "green"
		}

		return fmt.Sprintf("You've got %v more seasons ready to watch!", totalSeasons-watchedSeasons), "green"
	}

	return "Unknown", "grey"
}
