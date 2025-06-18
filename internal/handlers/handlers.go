package handlers

import (
	"database/sql"

	"AppDevelopmentAPI/internal/services"
)

type Handler struct {
	DB   *sql.DB
	QGen services.QuizGen
}

func New(db *sql.DB, key string) *Handler {
	return &Handler{DB: db, QGen: services.QuizGen{Key: key}}
}
