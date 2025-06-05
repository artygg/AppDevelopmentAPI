// Uncomment this code if you need websockets connection

package main

//
//import (
//	"log"
//	"net/url"
//	"os"
//	"os/signal"
//
//	"github.com/gorilla/websocket"
//)
//
//func main() {
//	interrupt := make(chan os.Signal, 1)
//	signal.Notify(interrupt, os.Interrupt)
//
//	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
//	log.Printf("Connecting to %s", u.String())
//
//	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
//	if err != nil {
//		log.Fatal("Dial error:", err)
//	}
//	defer conn.Close()
//
//	done := make(chan struct{})
//
//	// Listener goroutine
//	go func() {
//		defer close(done)
//		for {
//			_, message, err := conn.ReadMessage()
//			if err != nil {
//				log.Println("Read error:", err)
//				return
//			}
//			log.Printf("Received: %s", message)
//		}
//	}()
//
//	// Wait for interrupt (Ctrl+C)
//	<-interrupt
//	log.Println("Interrupt received, closing connection")
//
//	// Gracefully close
//	err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
//	if err != nil {
//		log.Println("Close message error:", err)
//		return
//	}
//
//	<-done
//}
