// internal/handlers/image.go
package handlers

import "net/http"

func (h *Handler) UploadImage(w http.ResponseWriter, _ *http.Request) {
    http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) GetImage(w http.ResponseWriter, _ *http.Request) {
    http.Error(w, "not implemented", http.StatusNotImplemented)
}

