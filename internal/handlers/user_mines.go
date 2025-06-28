// internal/handlers/user_mines.go
package handlers

import (
	"AppDevelopmentAPI/internal/models"
	"encoding/json"
	"net/http"

	"github.com/jmoiron/sqlx" // ADD
)

// GET /api/mines?user=john
func (h *Handler) GetMineBalance(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	if user == "" {
		http.Error(w, "missing user", http.StatusBadRequest)
		return
	}

	dbx := sqlx.NewDb(h.DB, "postgres")

	bal, err := models.GetMineBalance(dbx, user)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(struct {
		Balance int `json:"balance"`
	}{bal})
}
