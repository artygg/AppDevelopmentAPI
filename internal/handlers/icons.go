// internal/handlers/icons.go
package handlers

import (
    "encoding/json"
    "fmt"
    "net/http"
)

// GET /icon?category_id=3
func (h *Handler) IconLookup(w http.ResponseWriter, r *http.Request) {
    cid := r.URL.Query().Get("category_id")
    if cid == "" {
        http.Error(w, "missing id", http.StatusBadRequest)
        return
    }
    var name string
    err := h.DB.QueryRow(
        `SELECT icon_name FROM category_icons WHERE category_id = $1`, cid,
    ).Scan(&name)
    if err != nil {
        http.Error(w, "not found", http.StatusNotFound)
        return
    }
    _ = json.NewEncoder(w).Encode(name)
}

// GET /category_icons.json
func (h *Handler) CategoryIcons(w http.ResponseWriter, _ *http.Request) {
    rows, _ := h.DB.Query(`SELECT category_id, icon_name FROM category_icons`)
    defer rows.Close()

    m := map[string]string{}
    for rows.Next() {
        var id int
        var n  string
        _ = rows.Scan(&id, &n)
        m[fmt.Sprint(id)] = n
    }
    _ = json.NewEncoder(w).Encode(m)
}
