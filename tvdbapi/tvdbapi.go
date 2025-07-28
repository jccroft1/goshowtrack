package tvdbapi

import (
	"cmp"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"time"

	"github.com/jccroft1/goshowtrack/db"
)

const (
	baseUrl      string = "https://api.themoviedb.org/3/"
	baseImageURL string = "https://media.themoviedb.org/t/p/w300_and_h450_bestv2"
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
	limiter = time.Tick(80 * time.Millisecond)

	token string
)

func Setup(_token string) {
	token = _token

	go func() {
		t := time.Tick(time.Hour * 48)
		for range t {
			refreshShows()
		}
	}()
}

func refreshShows() {
	log.Println("Refreshing shows...")

	rows, err := db.Connection.Query("SELECT show_id FROM shows")
	if err != nil {
		log.Println("failed to load show ids", err)
		return
	}

	var ids []int
	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			log.Println("failed to scan show id", err)
			continue
		}

		ids = append(ids, id)
	}
	rows.Close()

	for _, id := range ids {
		fmt.Println("Refreshing show: ", id)
		_, err = GetShowDetails(id, true)
		if err != nil {
			log.Println("failed to get show details", err)
			continue
		}
	}
	log.Println("Refresh complete.")
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
	LastAirDate  string
}

type SeasonDetails struct {
	Episodes []Episode `json:"episodes"`
}

type Episode struct {
	AirDate string `json:"air_date"`
}

// https://developer.themoviedb.org/reference/tv-series-details
// https://developer.themoviedb.org/reference/tv-season-details
func GetShowDetails(id int, forceRefresh bool) (*ShowDetail, error) {
	if !forceRefresh {
		var show ShowDetail
		err := db.Connection.QueryRow("SELECT show_id, name, status, air_date, description, poster_path FROM shows WHERE show_id = ?", id).
			Scan(&show.ID, &show.Name, &show.Status, &show.AirDate, &show.Description, &show.PosterPath)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to check show details in DB: %v", err)
		}
		if err != sql.ErrNoRows {
			// load seasons
			rows, err := db.Connection.Query("SELECT name, episode_count, season_number, air_date, last_air_date FROM seasons WHERE show_id = ?", id)
			if err != nil {
				return nil, fmt.Errorf("failed to query seasons: %v", err)
			}
			defer rows.Close()

			seasons := []Season{}
			for rows.Next() {
				var season Season
				err := rows.Scan(&season.Name, &season.EpisodeCount, &season.Number, &season.AirDate, &season.LastAirDate)
				if err != nil {
					return nil, fmt.Errorf("failed to scan season: %v", err)
				}
				seasons = append(seasons, season)
			}

			show.Seasons = seasons

			return &show, err
		}
	}

	// actual request
	var response ShowDetail
	err := getRequest("tv/"+strconv.Itoa(id), &response)
	if err != nil {
		return nil, err
	}
	response.PosterPath = fmt.Sprintf("%s%s", baseImageURL, response.PosterPath)

	// remove Season 0 (specials)
	for i, s := range response.Seasons {
		if s.Number != 0 {
			continue
		}

		response.Seasons = append(response.Seasons[:i], response.Seasons[i+1:]...)
		break // assume there's only 1 season 0
	}

	// bit of a hack, we fetch each Season details, then find the newest episode and augment the response
	for i, s := range response.Seasons {
		var season SeasonDetails
		err = getRequest(fmt.Sprintf("tv/%v/season/%v", strconv.Itoa(id), s.Number), &season)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch season %v details for show %d: %v", s.Number, id, err)
		}

		if len(season.Episodes) == 0 {
			continue
		}

		newestEpisode := slices.MaxFunc(season.Episodes, func(a, b Episode) int {
			return cmp.Compare(a.AirDate, b.AirDate)
		})

		response.Seasons[i].LastAirDate = newestEpisode.AirDate
	}

	// save to DB
	query := `INSERT INTO shows (show_id, name, status, air_date, description, poster_path) 
	VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(show_id) DO UPDATE SET
		name = excluded.name, 
		status = excluded.status, 
		air_date = excluded.air_date, 
		description = excluded.description, 
		poster_path = excluded.poster_path;`
	_, err = db.Connection.Exec(query,
		response.ID, response.Name, response.Status, response.AirDate, response.Description, response.PosterPath)
	if err != nil {
		return &ShowDetail{}, fmt.Errorf("failed to insert show: %v", err)

	}

	for _, s := range response.Seasons {
		query = `INSERT INTO seasons (show_id, name, season_number, episode_count, air_date, last_air_date) 
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(show_id, season_number) DO UPDATE SET
			name = excluded.name, 
			episode_count = excluded.episode_count, 
			air_date = excluded.air_date, 
			last_air_date = excluded.last_air_date;`
		_, err = db.Connection.Exec(query,
			response.ID, s.Name, s.Number, s.EpisodeCount, s.AirDate, s.LastAirDate)
		if err != nil {
			return &ShowDetail{}, fmt.Errorf("failed to insert season: %v", err)
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
	<-limiter
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
