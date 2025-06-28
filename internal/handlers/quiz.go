package handlers

import (
	"AppDevelopmentAPI/internal/models"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

const quizTTL = 24 * time.Hour

func (h *Handler) Quiz(w http.ResponseWriter, r *http.Request) {
	var place *models.Place
	var err error

	if id := r.URL.Query().Get("place_id"); id != "" {
		pid, _ := strconv.Atoi(id)
		place, err = models.GetByID(h.DB, pid)
	} else if n := r.URL.Query().Get("place"); n != "" {
		place, err = models.GetByName(h.DB, n)
	}
	if err != nil || place == nil {
		http.Error(w, "place not found", http.StatusNotFound)
		return
	}

	if q, _ := models.GetByPlaceID(h.DB, place.ID); q != nil && time.Since(q.UpdatedAt) < quizTTL {
		json.NewEncoder(w).Encode(q)
		return
	}

	qs, err := h.QGen.Generate(place.Name, place.Latitude, place.Longitude)
	if err != nil || len(qs) != 7 {
		http.Error(w, "failed to generate quiz", http.StatusInternalServerError)
		return
	}

	newQuiz := models.Quiz{PlaceID: place.ID, Questions: qs}
	if err := models.Store(h.DB, place.ID, newQuiz); err != nil {
		log.Println("quiz store error:", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(newQuiz)
}
