package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (h *Handler) ListMines(w http.ResponseWriter, r *http.Request) {
	pid, _ := strconv.Atoi(r.URL.Query().Get("place_id"))
	if pid == 0 {
		http.Error(w, "missing place_id", http.StatusBadRequest)
		return
	}
	rows, err := h.DB.Query(`SELECT qid FROM mines
	                          WHERE place_id = $1 AND expires_at > NOW()`, pid)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var qid string
		_ = rows.Scan(&qid)
		ids = append(ids, qid)
	}
	_ = json.NewEncoder(w).Encode(struct {
		QIDs []string `json:"qids"`
	}{ids})
}
