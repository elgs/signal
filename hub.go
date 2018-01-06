package main

import (
	"fmt"
	"log"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	id  string
	pin string

	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	defer func() {
		delete(hubs, h.id)
		log.Printf("Hub %v closed", h.id)
	}()
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			client.send <- []byte(fmt.Sprintf(`{"type":"id", "id":"%v"}`, client.id))
			log.Printf("Client %v registered to hub %v", client.id, h.id)
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client %v unregistered from hub %v", client.id, h.id)
				go func() { h.broadcast <- []byte(fmt.Sprintf(`{"type":"leave", "from":"%v"}`, client.id)) }()
				if len(h.clients) == 0 {
					log.Printf("Hub %v is closing", h.id)
					return
				}
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
