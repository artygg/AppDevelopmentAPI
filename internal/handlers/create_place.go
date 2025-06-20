// internal/handlers/create_place.go
package handlers

import (
	"AppDevelopmentAPI/internal/models"
	"AppDevelopmentAPI/internal/services"
	"encoding/json"
	"net/http"
	"time"
)

// POST /api/places
func (h *Handler) CreatePlace(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var p models.Place
	if json.NewDecoder(r.Body).Decode(&p) != nil {
		http.Error(w, "bad body", http.StatusBadRequest)
		return
	}

	if err := h.DB.QueryRow(`
        INSERT INTO places(name, latitude, longitude, category_id, captured)
        VALUES ($1, $2, $3, $4, FALSE)
        RETURNING id`,
		p.Name, p.Latitude, p.Longitude, p.CategoryID).
		Scan(&p.ID); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	// Send update after place creation
	update := models.Update{
		Status:    "created",
		Time:      time.Now().Format(time.RFC3339),
		Source:    "-",
		PlaceID:   p.ID,
		PlaceName: p.Name,
	}
	services.SendUpdate(update)

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(p)
}
