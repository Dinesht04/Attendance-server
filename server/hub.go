package server

import (
	"sync"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan WsEvent
	id   string
	role string
}

type Message struct {
	ClientID string
	Type     string
	Text     WsEvent
}

type Hub struct {
	sync.RWMutex
	Clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
	db         *mongo.Client
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.Clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.Clients[client]; ok {
				close(client.send)
				delete(h.Clients, client)
			}
		case msg := <-h.broadcast:
			for client, connected := range h.Clients {
				if connected {
					client.send <- msg.Text
				}
			}
		}
	}
}
