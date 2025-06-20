package models

type Update struct {
	Status    string `json:"status"`
	Time      string `json:"time"`
	Source    string `json:"source"`
	PlaceID   int    `json:"place_id,omitempty"`
	PlaceName string `json:"place_name,omitempty"`
}
