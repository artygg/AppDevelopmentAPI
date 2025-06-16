package main

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func UploadImageHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid method. Use POST.", http.StatusMethodNotAllowed)
			return
		}

		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
			return
		}

		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Failed to read file: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		placeID := r.FormValue("place_id")
		if placeID == "" {
			http.Error(w, "Missing place_id in form data", http.StatusBadRequest)
			return
		}

		// Ensure directory exists
		if _, err := os.Stat("images"); os.IsNotExist(err) {
			if err := os.Mkdir("images", 0755); err != nil {
				http.Error(w, "Failed to create images directory", http.StatusInternalServerError)
				return
			}
		}

		timestamp := time.Now().Format("20060102_150405")
		filename := fmt.Sprintf("%s_%s", timestamp, filepath.Base(handler.Filename))
		fullPath := filepath.Join("images", filename)

		dst, err := os.Create(fullPath)
		if err != nil {
			http.Error(w, "Failed to create file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, "Failed to save file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		now := time.Now()

		query := `
			INSERT INTO public.image_place (image_location, place_id, last_time_updated)
			VALUES ($1, $2, $3)
			ON CONFLICT (place_id)
			DO UPDATE SET
				image_location = EXCLUDED.image_location,
				last_time_updated = EXCLUDED.last_time_updated;
		`

		_, err = db.Exec(query, fullPath, placeID, now)
		if err != nil {
			http.Error(w, "Database insert/update failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "File uploaded and saved to DB as %s", filename)
	}
}

func GetImageByPlaceIDHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		placeID := r.URL.Query().Get("place_id")
		if placeID == "" {
			http.Error(w, "Missing place_id query parameter", http.StatusBadRequest)
			return
		}

		var imagePath string
		query := `SELECT image_location FROM public.image_place WHERE place_id = $1`
		err := db.QueryRow(query, placeID).Scan(&imagePath)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Image not found", http.StatusNotFound)
			} else {
				http.Error(w, "Database query error: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		file, err := os.Open(imagePath)
		if err != nil {
			http.Error(w, "Failed to open image: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		w.Header().Set("Content-Type", "image/jpeg")
		io.Copy(w, file)
	}
}
