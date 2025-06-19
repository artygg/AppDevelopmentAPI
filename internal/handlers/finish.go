
package handlers

import (
    "AppDevelopmentAPI/internal/models"
    "encoding/json"
    "net/http"
    "time"
)

const minCorrectToCapture = 1

// DTO
type finishReq struct {
    PlaceID   int    `json:"place_id"`
    User      string `json:"user"`
    Correct   int    `json:"correct"`
    ElapsedMs int64  `json:"elapsed_ms"`
}
type finishResp struct {
    Captured bool          `json:"captured"`
    Quiz     *models.Quiz  `json:"quiz,omitempty"`
}


// POST /api/finish
func (h *Handler) Finish(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }


    var req finishReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "bad body", http.StatusBadRequest)
        return
    }


    _, _ = h.DB.Exec(`
        INSERT INTO capture_attempts(place_id,user_name,correct,time_ms)
        VALUES ($1,$2,$3,$4)`,
        req.PlaceID, req.User, req.Correct, req.ElapsedMs)


    _, _ = h.DB.Exec(`
        INSERT INTO place_scores (place_id,best_correct,best_time_ms,holder)
        VALUES ($1,$2,$3,$4)
        ON CONFLICT (place_id) DO UPDATE
        SET best_correct = EXCLUDED.best_correct,
            best_time_ms = EXCLUDED.best_time_ms,
            holder       = EXCLUDED.holder,
            updated_at   = now()
        WHERE EXCLUDED.best_correct >  place_scores.best_correct
           OR (EXCLUDED.best_correct = place_scores.best_correct
               AND EXCLUDED.best_time_ms < place_scores.best_time_ms)`,
        req.PlaceID, req.Correct, req.ElapsedMs, req.User)

 
    var holder bool
    _ = h.DB.QueryRow(`
        SELECT holder = $1
        FROM   place_scores
        WHERE  place_id = $2`,
        req.User, req.PlaceID).Scan(&holder)

    captured := holder && req.Correct >= minCorrectToCapture


    var newQuiz *models.Quiz
    if captured {
        // 5-a) mark as captured
        _, _ = h.DB.Exec(`
            UPDATE places
            SET captured      = TRUE,
                user_captured = $1,
                captured_at   = $2
            WHERE id = $3`,
            req.User, time.Now(), req.PlaceID)


        if p, err := models.GetByID(h.DB, req.PlaceID); err == nil {
            if qs, err2 := h.QGen.Generate(p.Name, p.Latitude, p.Longitude); err2 == nil && len(qs) == 7 {
                q := models.Quiz{PlaceID: req.PlaceID, Questions: qs}
                _ = models.Store(h.DB, req.PlaceID, q)
                newQuiz = &q
            }
        }
    }


    _ = json.NewEncoder(w).Encode(finishResp{
        Captured: captured,
        Quiz:     newQuiz,
    })
}
