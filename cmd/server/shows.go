package main

type Show struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Year        string `json:"year"`
	Description string `json:"description"`
	CoverURL    string `json:"cover_url"`
}

type ListShowsResponse struct {
	Shows []Show `json:"shows"`
}

// func searchShowsHandler(w http.ResponseWriter, r *http.Request) {
// 	var req struct {
// 		Query string `json:"query"`
// 	}
// 	err := json.NewDecoder(r.Body).Decode(&req)
// 	if err != nil || req.Query == "" {
// 		log.Println("Query:", req.Query, err)
// 		http.Error(w, "Invalid query", http.StatusBadRequest)
// 		return
// 	}

// 	searchResults, err := tvdbapi.SearchShow(req.Query)
// 	if err != nil {
// 		log.Println("Error searching shows:", err)
// 		http.Error(w, "Failed to search shows", http.StatusInternalServerError)
// 		return
// 	}

// 	response := ListShowsResponse{}
// 	for _, show := range searchResults {
// 		year := ""
// 		if len(show.AirDate) > 4 {
// 			year = show.AirDate[0:4]
// 		}

// 		response.Shows = append(response.Shows, Show{
// 			ID:          show.ID,
// 			Name:        show.Name,
// 			Year:        year,
// 			Description: show.Description,
// 			CoverURL:    show.PosterPath,
// 		})
// 	}

// 	json.NewEncoder(w).Encode(response)
// }

type Season struct {
	Number       int    `json:"season_number"`
	Name         string `json:"name"`
	EpisodeCount int    `json:"episode_count"`
}

type AddShowResponse struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Year        string   `json:"year"`
	Description string   `json:"description"`
	CoverURL    string   `json:"cover_url"`
	Seasons     []Season `json:"seasons"`
}

// func addShowHandler(w http.ResponseWriter, r *http.Request) {
// 	var req struct {
// 		ID int `json:"id"`
// 	}
// 	err := json.NewDecoder(r.Body).Decode(&req)
// 	if err != nil || req.ID == 0 {
// 		log.Println("ID:", req.ID, err)
// 		http.Error(w, "Invalid ID", http.StatusBadRequest)
// 		return
// 	}

// 	// add show to user's list of shows in the DB

// 	// fetch the details for the show
// 	showDetails, err := tvdbapi.GetShowDetails(req.ID)
// 	if err != nil {
// 		log.Println("Error fetching show details:", err)
// 		http.Error(w, "Failed to fetch show details", http.StatusInternalServerError)
// 		return
// 	}

// 	year := ""
// 	if len(showDetails.AirDate) > 4 {
// 		year = showDetails.AirDate[0:4]
// 	}
// 	var responseSeasons []Season
// 	for _, season := range showDetails.Seasons {
// 		if season.Number == 0 {
// 			continue
// 		}
// 		responseSeasons = append(responseSeasons, Season{
// 			Number:       season.Number,
// 			Name:         season.Name,
// 			EpisodeCount: season.EpisodeCount,
// 		})
// 	}
// 	response := AddShowResponse{
// 		ID:          showDetails.ID,
// 		Name:        showDetails.Name,
// 		Year:        year,
// 		Description: showDetails.Description,
// 		CoverURL:    showDetails.PosterPath,
// 		Seasons:     responseSeasons,
// 	}
// 	json.NewEncoder(w).Encode(response)
// }

// func listShowsHandler(w http.ResponseWriter, r *http.Request) {
// 	email, ok := getUserEmail(r)
// 	if !ok {
// 		log.Panicln("Failed to get user email")
// 		http.Error(w, "Invalid email", http.StatusUnauthorized)
// 		return
// 	}

// 	fmt.Println(email)

// 	response := ListShowsResponse{
// 		Shows: []Show{
// 			{ID: 1, Name: "The Office"},
// 			{ID: 2, Name: "Breaking Bad"},
// 		},
// 	}

// 	json.NewEncoder(w).Encode(response)
// }
