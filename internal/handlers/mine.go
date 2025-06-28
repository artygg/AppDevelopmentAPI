// internal/handlers/mine.go
package handlers

import (
	"AppDevelopmentAPI/internal/models" // only used for GetByPlaceID – keep it
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

func (h *Handler) Mine(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// ──────────────────────────────── INPUT ────────────────────────────────
	var req struct {
		PlaceID int    `json:"place_id"`
		QID     string `json:"qid"`
		User    string `json:"user"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.User == "" {
		http.Error(w, "bad body", http.StatusBadRequest)
		return
	}

	// ──────────────────────── ENSURE BALANCE ROW EXISTS ───────────────────────
	_, _ = h.DB.Exec(`INSERT INTO user_mines(username,balance)
	                  VALUES ($1,0)
	                  ON CONFLICT (user_name) DO NOTHING`, req.User)

	// ────────────────────────── READ CURRENT BALANCE ──────────────────────────
	var balance int
	if err := h.DB.QueryRow(
		`SELECT balance FROM user_mines WHERE username = $1`, req.User,
	).Scan(&balance); err != nil && err != sql.ErrNoRows {
		log.Println("balance query error:", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	if balance <= 0 {
		http.Error(w, "no mines left", http.StatusForbidden)
		return
	}

	// ─────────────────────────── DEDUCT ONE MINE ─────────────────────────────
	if _, err := h.DB.Exec(
		`UPDATE user_mines SET balance = balance - 1
		  WHERE username = $1 AND balance > 0`, req.User); err != nil {
		log.Println("consume mine error:", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	balance-- // reflect locally

	// ─────────────────────────── STORE / REFRESH MINE ─────────────────────────
	if _, err := h.DB.Exec(`
		INSERT INTO mines(place_id,qid,expires_at)
		VALUES ($1,$2,now() + interval '24 hours')
		ON CONFLICT (place_id,qid) DO
		     UPDATE SET expires_at = now() + interval '24 hours'`,
		req.PlaceID, req.QID); err != nil {

		// rollback deduction best-effort
		_, _ = h.DB.Exec(`UPDATE user_mines SET balance = balance + 1
		                  WHERE user_name = $1`, req.User)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	// ───────────── PATCH THE QUESTION’s timeLimit → 5 s (non-fatal on error) ──
	_, _ = h.DB.Exec(`
		WITH target AS (
		    SELECT idx - 1 AS i
		    FROM   jsonb_array_elements(quiz_json->'questions')
		           WITH ORDINALITY q(obj,idx)
		    WHERE  obj ->> 'id' = $2
		)
		UPDATE quizzes
		SET    quiz_json = jsonb_set(
		           quiz_json,
		           '{questions,' || (SELECT i FROM target) || ',timeLimit}',
		           to_jsonb(5),
		           true)
		WHERE  place_id = $1`,
		req.PlaceID, req.QID)

	// ──────────────────────────── BUILD RESPONSE ─────────────────────────────
	quiz, _ := models.GetByPlaceID(h.DB, req.PlaceID)

	var resp struct {
		Quiz        *models.Quiz `json:"quiz"`
		MineBalance int          `json:"mine_balance"`
	}
	resp.Quiz = quiz
	resp.MineBalance = balance

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201
	_ = json.NewEncoder(w).Encode(resp)
}
