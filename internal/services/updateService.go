package services

import (
	"AppDevelopmentAPI/internal/models"
	"AppDevelopmentAPI/websocket"
	"encoding/json"
)

func SendUpdate(u models.Update) {
	if b, err := json.Marshal(u); err == nil {
		websocket.Broadcast <- b
	}
}
