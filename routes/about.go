package routes

import (
	"net/http"

	"github.com/jccroft1/goshowtrack/auth"
)

func AboutHandler(w http.ResponseWriter, req *http.Request) {
	email, ok := auth.GetUserEmail(req)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	type AboutData struct {
		Email string
	}

	data := AboutData{
		Email: email,
	}

	renderTemplate(w, "about", data)
}
