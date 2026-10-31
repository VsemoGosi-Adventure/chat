package websocket

import (
	"context"
	"sync"

	"github.com/VsemoGosi-Adventure/chat/internal/domain"
	"github.com/google/uuid"
)

type Hub struct {
	rooms          map[uuid.UUID]*Room
	register       chan *Client
	unregister     chan *Client
	broadcast      chan *BroadcastMessage
	mu             sync.RWMutex
	broadcastBuffer int
}

type Room struct {
	eventID uuid.UUID
	clients map[*Client]bool
	mu      sync.RWMutex
}

type BroadcastMessage struct {
	EventID uuid.UUID
	Message domain.OutgoingMessage
}

func NewHub(broadcastBuffer int) *Hub {
	return &Hub{
		rooms:          make(map[uuid.UUID]*Room),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		broadcast:      make(chan *BroadcastMessage, broadcastBuffer),
		broadcastBuffer: broadcastBuffer,
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			h.shutdown()
			return
		case client := <-h.register:
			h.registerClient(client)
		case client := <-h.unregister:
			h.unregisterClient(client)
		case msg := <-h.broadcast:
			h.broadcastToRoom(msg)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, exists := h.rooms[client.eventID]
	if !exists {
		room = &Room{
			eventID: client.eventID,
			clients: make(map[*Client]bool),
		}
		h.rooms[client.eventID] = room
	}

	room.mu.Lock()
	room.clients[client] = true
	room.mu.Unlock()
}

func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, exists := h.rooms[client.eventID]
	if !exists {
		return
	}

	room.mu.Lock()
	if _, ok := room.clients[client]; ok {
		delete(room.clients, client)
		close(client.send)
	}
	room.mu.Unlock()

	if len(room.clients) == 0 {
		delete(h.rooms, client.eventID)
	}
}

func (h *Hub) broadcastToRoom(msg *BroadcastMessage) {
	h.mu.RLock()
	room, exists := h.rooms[msg.EventID]
	h.mu.RUnlock()

	if !exists {
		return
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	for client := range room.clients {
		select {
		case client.send <- msg.Message:
		default:
			go func(c *Client) {
				h.unregister <- c
			}(client)
		}
	}
}

func (h *Hub) Broadcast(eventID uuid.UUID, msg domain.OutgoingMessage) {
	h.broadcast <- &BroadcastMessage{
		EventID: eventID,
		Message: msg,
	}
}

func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

func (h *Hub) shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, room := range h.rooms {
		room.mu.Lock()
		for client := range room.clients {
			close(client.send)
		}
		room.mu.Unlock()
	}

	h.rooms = make(map[uuid.UUID]*Room)
}
