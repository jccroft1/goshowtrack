package tvdbapi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gotrack/db"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Show struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	AirDate     string `json:"first_air_date"`
	Description string `json:"overview"`
	PosterPath  string `json:"poster_path"`
}

type SearchShowsResponse struct {
	Results []Show `json:"results"`
}

var (
	client = &http.Client{
		Timeout: time.Second,
	}

	baseUrl string
	token   string

	baseImageURL string = "https://media.themoviedb.org/t/p/w300_and_h450_bestv2"
)

func Setup(_baseUrl string, _token string) {
	baseUrl = _baseUrl
	token = _token
}

func SearchShow(query string) ([]Show, error) {
	escapedQuery := url.QueryEscape(query)

	url := fmt.Sprintf("search/tv?query=%v&include_adult=false&language=en-US&page=1", escapedQuery)
	var results SearchShowsResponse
	err := getRequest(url, &results)
	if err != nil {
		return nil, err
	}

	for i, show := range results.Results {
		show.PosterPath = fmt.Sprintf("%s%s", baseImageURL, show.PosterPath)
		results.Results[i] = show
	}

	return results.Results, nil
}

type ShowDetail struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Status      string   `json:"status"`
	AirDate     string   `json:"first_air_date"`
	Description string   `json:"overview"`
	PosterPath  string   `json:"poster_path"`
	Seasons     []Season `json:"seasons"`
}

type Season struct {
	Number       int    `json:"season_number"`
	Name         string `json:"name"`
	EpisodeCount int    `json:"episode_count"`
	AirDate      string `json:"air_date"`
}

// https://api.themoviedb.org/3/tv/253?language=en-US'
func GetShowDetails(id int) (*ShowDetail, error) {
	// attempt to load from DB
	var show ShowDetail
	fmt.Println("ID", id)
	err := db.Connection.QueryRow("SELECT show_id, name, status, air_date, description, poster_path FROM shows WHERE show_id = ?", id).
		Scan(&show.ID, &show.Name, &show.Status, &show.AirDate, &show.Description, &show.PosterPath)
	if err != sql.ErrNoRows {
		log.Println("loaded show from cache")

		// load seasons
		rows, err := db.Connection.Query("SELECT name, episode_count, season_number, air_date FROM seasons WHERE show_id = ?", id)
		if err != nil {
			return nil, fmt.Errorf("failed to query seasons: %v", err)
		}
		defer rows.Close()

		seasons := []Season{}
		for rows.Next() {
			var season Season
			err := rows.Scan(&season.Name, &season.EpisodeCount, &season.Number, &season.AirDate)
			if err != nil {
				return nil, fmt.Errorf("failed to scan season: %v", err)
			}
			seasons = append(seasons, season)
		}

		show.Seasons = seasons

		return &show, err
	}

	var response ShowDetail

	err = getRequest("tv/"+strconv.Itoa(id), &response)
	if err != nil {
		return nil, err
	}
	response.PosterPath = fmt.Sprintf("%s%s", baseImageURL, response.PosterPath)

	// save to DB
	_, err = db.Connection.Exec(`INSERT INTO shows (show_id, name, status, air_date, description, poster_path) VALUES (?, ?, ?, ?, ?, ?);`,
		response.ID, response.Name, response.Status, response.AirDate, response.Description, response.PosterPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, s := range response.Seasons {
		_, err = db.Connection.Exec(`INSERT INTO seasons (show_id, name, season_number, episode_count, air_date) VALUES (?, ?, ?, ?, ?);`,
			response.ID, s.Name, s.Number, s.EpisodeCount, s.AirDate)
		if err != nil {
			log.Fatal(err)
		}
	}

	return &response, nil
}

func getRequest(relativeURL string, output interface{}) error {
	url := fmt.Sprintf("%s%s", baseUrl, relativeURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	fmt.Println(req.URL)
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	err = json.NewDecoder(res.Body).Decode(&output)
	if err != nil {
		return err
	}

	return nil
}
