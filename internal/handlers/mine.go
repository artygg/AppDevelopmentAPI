// internal/handlers/mine.go
package handlers

import (
    "encoding/json"
    "log"
    "net/http"

    "AppDevelopmentAPI/internal/models"
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
    if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
        http.Error(w, "bad body", http.StatusBadRequest)
        return
    }


    if _, err := h.DB.Exec(`
        INSERT INTO mines(place_id,qid,expires_at)
        VALUES ($1,$2,now() + interval '24 hours')
        ON CONFLICT (place_id,qid) DO UPDATE
              SET expires_at = now() + interval '24 hours'`,
        m.PlaceID, m.QID); err != nil {

        http.Error(w, "db error", http.StatusInternalServerError)
        return
    }


    if _, err := h.DB.Exec(`
        UPDATE quizzes
        SET  quiz_json = jsonb_set(
                 quiz_json,
                 '{questions,' ||
                   (SELECT idx-1
                    FROM   jsonb_array_elements(quiz_json->'questions')
                           WITH ORDINALITY q(obj,idx)
                    WHERE  obj->>'id' = $2)::text ||
                 ',timeLimit}',
                 to_jsonb(5),
                 true)
        WHERE place_id = $1`,
        m.PlaceID, m.QID); err != nil {

        // не фатальная ошибка – просто логируем
        log.Println("quiz patch error:", err)
    }

    patched, err := models.GetByPlaceID(h.DB, m.PlaceID)
    if err != nil {
        log.Println("quiz fetch error:", err)
        w.WriteHeader(http.StatusCreated) // всё-равно 201
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)    // 201
    _ = json.NewEncoder(w).Encode(patched)
}
