package server

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan WsReq
	id   string
}

type Message struct {
	ClientID string
	Text     WsReq
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
			fmt.Println("client registered: ", client.id)

		case client := <-h.unregister:
			if _, ok := h.Clients[client]; ok {
				close(client.send)
				delete(h.Clients, client)
			}
		case msg := <-h.broadcast:
			fmt.Println("mongo db call")
			fmt.Println("broadcast this to everyone")
			fmt.Println(h.Clients)
			for client, connected := range h.Clients {
				fmt.Println(client.id)
				if connected {
					client.send <- msg.Text
				}
			}
		}
	}
}
