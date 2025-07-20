package routes

import (
	"database/sql"
	"gotrack/auth"
	"gotrack/db"
	"gotrack/mail"
	"log"
	"net/http"
	"time"
)

func RootHandler(w http.ResponseWriter, req *http.Request) {
	_, _, ok := auth.Validate(req)
	if ok {
		http.Redirect(w, req, "/home", http.StatusFound)
		return
	}

	renderTemplate(w, "login", nil)
}

func LoginHandler(w http.ResponseWriter, req *http.Request) {
	email := req.FormValue("email")
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	token, err := auth.GenerateAndStoreToken(email)
	if err != nil {
		log.Println(err)
		http.Error(w, "Could not create token", http.StatusInternalServerError)
		return
	}

	err = mail.SendMagicLink(email, token)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	// Render the login template or redirect to login page
	renderTemplate(w, "loginRedirect", nil)
}

func AuthenticateHandler(w http.ResponseWriter, req *http.Request) {
	// read token from query string
	token := req.URL.Query().Get("token")
	if token == "" {
		http.Redirect(w, req, "/", http.StatusSeeOther)
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
