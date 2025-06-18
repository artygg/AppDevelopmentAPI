// internal/handlers/mine.go
package handlers

import (
    "encoding/json"
    "net/http"
)

func (h *Handler) Mine(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    var m struct {
        PlaceID int    `json:"place_id"`
        QID     string `json:"qid"`
    }
    if json.NewDecoder(r.Body).Decode(&m) != nil {
        http.Error(w, "bad body", http.StatusBadRequest)
        return
    }
    if _, err := h.DB.Exec(`
        INSERT INTO mines(place_id,qid,expires_at)
        VALUES ($1,$2,now()+interval '24 hours')
        ON CONFLICT(place_id,qid) DO UPDATE
          SET expires_at = now()+interval '24 hours'`,
        m.PlaceID, m.QID); err != nil {
        http.Error(w, "db error", http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusCreated)
}

