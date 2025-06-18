// internal/handlers/finish.go
package handlers

import (
    "encoding/json"
    "net/http"
)

func (h *Handler) Finish(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    var f struct {
        PlaceID int    `json:"place_id"`
        User    string `json:"user"`
        Correct int    `json:"correct"`
        TimeMs  int64  `json:"elapsed_ms"`
    }
    if json.NewDecoder(r.Body).Decode(&f) != nil {
        http.Error(w, "bad body", http.StatusBadRequest)
        return
    }

    h.DB.Exec(`INSERT INTO capture_attempts(place_id,user_name,correct,time_ms)
               VALUES($1,$2,$3,$4)`,
        f.PlaceID, f.User, f.Correct, f.TimeMs)

    h.DB.Exec(`
      INSERT INTO place_scores(place_id,best_correct,best_time_ms,holder)
      VALUES($1,$2,$3,$4)
      ON CONFLICT (place_id) DO UPDATE
      SET best_correct = EXCLUDED.best_correct,
          best_time_ms = EXCLUDED.best_time_ms,
          holder       = EXCLUDED.holder,
          updated_at   = now()
      WHERE EXCLUDED.best_correct > place_scores.best_correct
         OR (EXCLUDED.best_correct = place_scores.best_correct
             AND EXCLUDED.best_time_ms < place_scores.best_time_ms)`,
        f.PlaceID, f.Correct, f.TimeMs, f.User)

    var captured bool
    _ = h.DB.QueryRow(`SELECT holder=$1 FROM place_scores WHERE place_id=$2`,
        f.User, f.PlaceID).Scan(&captured)

    json.NewEncoder(w).Encode(struct {
        Captured bool `json:"captured"`
    }{captured})
}

