// main.go
package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Place struct {
	Name       string     `json:"name"`
	Coordinate Coordinate `json:"coordinate"`
}

type Coordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func main() {
	places := []Place{
		{Name: "Wildlands Adventure Zoo", Coordinate: Coordinate{Latitude: 52.780748, Longitude: 6.887516}},
		{Name: "Station Emmen", Coordinate: Coordinate{Latitude: 52.790453, Longitude: 6.899715}},
		{Name: "Rensenpark", Coordinate: Coordinate{Latitude: 52.785692, Longitude: 6.897980}},
		{Name: "Emmerdennen Bos", Coordinate: Coordinate{Latitude: 52.794587, Longitude: 6.917414}},
		{Name: "Winkelcentrum De Weiert", Coordinate: Coordinate{Latitude: 52.782382, Longitude: 6.894363}},
		{Name: "NHL Stenden Emmen", Coordinate: Coordinate{Latitude: 52.778150, Longitude: 6.911960}},
		{Name: "Danackers 70", Coordinate: Coordinate{Latitude: 52.780455, Longitude: 6.942720}},
	}

	http.HandleFunc("/places", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(places); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
