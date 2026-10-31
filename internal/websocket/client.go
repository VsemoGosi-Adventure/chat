package websocket

import (
	"context"
	"encoding/json"
	"time"

	"github.com/VsemoGosi-Adventure/chat/internal/domain"
	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

type Client struct {
	hub            *Hub
	conn           *websocket.Conn
	send           chan domain.OutgoingMessage
	eventID        uuid.UUID
	userID         uuid.UUID
	pingInterval   time.Duration
	pongWait       time.Duration
	writeWait      time.Duration
	maxMessageSize int64
}

type ClientConfig struct {
	Hub            *Hub
	Conn           *websocket.Conn
	EventID        uuid.UUID
	UserID         uuid.UUID
	PingInterval   time.Duration
	PongWait       time.Duration
	WriteWait      time.Duration
	MaxMessageSize int64
	SendBuffer     int
}

func NewClient(cfg ClientConfig) *Client {
	return &Client{
		hub:            cfg.Hub,
		conn:           cfg.Conn,
		send:           make(chan domain.OutgoingMessage, cfg.SendBuffer),
		eventID:        cfg.EventID,
		userID:         cfg.UserID,
		pingInterval:   cfg.PingInterval,
		pongWait:       cfg.PongWait,
		writeWait:      cfg.WriteWait,
		maxMessageSize: cfg.MaxMessageSize,
	}
}

func (c *Client) ReadPump(ctx context.Context, onMessage func(context.Context, uuid.UUID, domain.IncomingMessage) error) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close(websocket.StatusNormalClosure, "")
	}()

	c.conn.SetReadLimit(c.maxMessageSize)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, data, err := c.conn.Read(ctx)
			if err != nil {
				return
			}

			var incoming domain.IncomingMessage
			if err := json.Unmarshal(data, &incoming); err != nil {
				continue
			}

			if err := onMessage(ctx, c.userID, incoming); err != nil {
				continue
			}
		}
	}
}

func (c *Client) WritePump(ctx context.Context) {
	ticker := time.NewTicker(c.pingInterval)
	defer func() {
		ticker.Stop()
		c.conn.Close(websocket.StatusNormalClosure, "")
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case message, ok := <-c.send:
			writeCtx, cancel := context.WithTimeout(ctx, c.writeWait)
			defer cancel()

			if !ok {
				c.conn.Close(websocket.StatusNormalClosure, "")
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				continue
			}

			if err := c.conn.Write(writeCtx, websocket.MessageText, data); err != nil {
				return
			}

		case <-ticker.C:
			writeCtx, cancel := context.WithTimeout(ctx, c.writeWait)
			defer cancel()

			if err := c.conn.Ping(writeCtx); err != nil {
				return
			}
		}
	}
}
