package websocket

import (
	"github.com/gorilla/websocket"
)

type WebSocketService struct {
	Upgrader *websocket.Upgrader
	Clients  map[*websocket.Conn]bool
	//TODO modify chan type if needed
	Broadcast chan string
}

func NewWebSocketService() *WebSocketService {
	service := WebSocketService{
		Upgrader:  &websocket.Upgrader{},
		Clients:   map[*websocket.Conn]bool{},
		Broadcast: make(chan string),
	}

	go service.Run()

	return &service
}

func (s *WebSocketService) Run() error {
	for {
		msg := <-s.Broadcast
		for client := range s.Clients {
			if err := client.WriteJSON(msg); err != nil {
				client.Close()
				// remove invalid client
				delete(s.Clients, client)
				return err
			}
		}
	}
	return nil
}
