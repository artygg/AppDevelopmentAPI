package handlers

import (
	"AppDevelopmentAPI/internal/models"
	"AppDevelopmentAPI/internal/services"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type captureReq struct {
	PlaceID int    `json:"place_id"`
	User    string `json:"user"`
	Passed  bool   `json:"passed"`
}

type captureResp struct {
	Place *models.Place `json:"place"`
	Quiz  *models.Quiz  `json:"quiz,omitempty"`
}

func (h *Handler) Capture(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req captureReq
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		http.Error(w, "bad body", http.StatusBadRequest)
		return
	}

	if req.Passed {
		_, err := h.DB.Exec(`
            UPDATE places
            SET captured = TRUE,
                user_captured = $1,
                captured_at = NOW()
            WHERE id = $2`, req.User, req.PlaceID)
		if err != nil {
			log.Println("db error 1 ", err)
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		h.DB.Exec(`DELETE FROM place_cooldowns
                   WHERE place_id = $1 AND user_name = $2`,
			req.PlaceID, req.User)
		_ = models.IncCaptured(h.DB, req.User)

		// Send update after successful capture
		place, _ := models.GetByID(h.DB, req.PlaceID)
		if place != nil {
			update := models.Update{
				Status:    "captured",
				Time:      time.Now().Format(time.RFC3339),
				Source:    req.User,
				PlaceID:   place.ID,
				PlaceName: place.Name,
			}
			services.SendUpdate(update)
		}
	} else {
		_, err := h.DB.Exec(`
            INSERT INTO place_cooldowns(place_id, user_name, cooldown_until)
            VALUES ($1, $2, NOW() + INTERVAL '24 hours')
            ON CONFLICT (place_id, user_name) DO UPDATE
              SET cooldown_until = EXCLUDED.cooldown_until`,
			req.PlaceID, req.User)
		if err != nil {
			log.Println("db error 2 ", err)
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
	}

	place, _ := models.GetByID(h.DB, req.PlaceID)

	var q *models.Quiz
	if req.Passed {
		qs, _ := h.QGen.Generate(place.Name, place.Latitude, place.Longitude)
		newQ := models.Quiz{PlaceID: place.ID, Questions: qs}
		_ = models.Store(h.DB, place.ID, newQ)
		q = &newQ
	}

	_ = json.NewEncoder(w).Encode(captureResp{place, q})
}
