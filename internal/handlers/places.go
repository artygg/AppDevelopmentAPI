package handlers

import (
	"AppDevelopmentAPI/internal/models"
	"encoding/json"
	"log"
	"net/http"
)

func (h *Handler) Places(w http.ResponseWriter, r *http.Request) {
	user := r.Header.Get("X-Player")
	if user == "" {
		user = "-"
	}

	list, err := models.GetAll(h.DB, user)
	if err != nil {
		log.Println(err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(list)
}
