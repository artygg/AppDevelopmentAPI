package main

import (
	"log"
	"net/http"
	"os"

	"AppDevelopmentAPI/internal/db"
	"AppDevelopmentAPI/internal/handlers"
	"AppDevelopmentAPI/websocket"
)

func main() {
	d := db.Connect()
	defer d.Close()

	h := handlers.New(d, os.Getenv("OPENAI_API_KEY"))

	mux := http.NewServeMux()

	mux.HandleFunc("/places",               h.Places)
	mux.HandleFunc("/quiz",                 h.Quiz)
	mux.HandleFunc("/icon",                 h.IconLookup)
	mux.HandleFunc("/category_icons.json",  h.CategoryIcons)

	mux.HandleFunc("/api/places",  h.CreatePlace)
	mux.HandleFunc("/api/capture", h.Capture)
	mux.HandleFunc("/api/mine",    h.Mine)
	mux.HandleFunc("/api/finish",  h.Finish)

	mux.HandleFunc("/upload-file", h.UploadImage)
	mux.HandleFunc("/get-image",   h.GetImage)

	go websocket.HandleMessages()
	mux.HandleFunc("/ws", websocket.WebSocketHandler)

	mux.Handle("/", http.FileServer(http.Dir(".")))

	log.Println("API listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
