package routes

import (
	"net/http"
	"os"
	"time"
)

func FaviconHandler(w http.ResponseWriter, r *http.Request) {
	// Define the path to your favicon file
	path := "../../assets/icon.png"

	// Open the file
	file, err := os.Open(path)
	if err != nil {
		http.Error(w, "Favicon not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Set the content type to PNG
	w.Header().Set("Content-Type", "image/png")

	// Serve the file
	http.ServeContent(w, r, "favicon.png", fileStat(file), file)
}

func fileStat(file *os.File) (modTime time.Time) {
	info, err := file.Stat()
	if err != nil {
		return
	}
	return info.ModTime()
}
