package handlers

import (
	"AppDevelopmentAPI/internal/models"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func (h *Handler) Leaderboard(w http.ResponseWriter, r *http.Request) {
	top := 100
	if s := r.URL.Query().Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			top = n
		}
	}

	list, err := models.Leaderboard(h.DB, top)
	if err != nil {
		log.Println(err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(list)
}
