package handler

import (
	"context"
	"net/http"

	"github.com/VsemoGosi-Adventure/chat/internal/config"
	"github.com/VsemoGosi-Adventure/chat/internal/domain"
	"github.com/VsemoGosi-Adventure/chat/internal/middleware"
	"github.com/VsemoGosi-Adventure/chat/internal/service"
	ws "github.com/VsemoGosi-Adventure/chat/internal/websocket"
	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

type WebSocketHandler struct {
	hub         *ws.Hub
	chatService *service.ChatService
	wsCfg       config.WebSocketConfig
}

func NewWebSocketHandler(hub *ws.Hub, chatService *service.ChatService, wsCfg config.WebSocketConfig) *WebSocketHandler {
	return &WebSocketHandler{
		hub:         hub,
		chatService: chatService,
		wsCfg:       wsCfg,
	}
}

// HandleWebSocket godoc
// @Summary      WebSocket connection for real-time chat
// @Description  Upgrade HTTP connection to WebSocket for real-time bidirectional messaging within an event chat
// @Tags         websocket
// @Accept       json
// @Produce      json
// @Param        event_id  path  string  true  "Event ID (UUID)"
// @Success      101  {string}  string "Switching Protocols"
// @Failure      400  {string}  string "Bad Request"
// @Failure      401  {string}  string "Unauthorized"
// @Failure      500  {string}  string "Internal Server Error"
// @Security     ApiKeyAuth
// @Router       /chats/{event_id}/ws [get]
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	eventIDStr := r.PathValue("event_id")
	if eventIDStr == "" {
		http.Error(w, "Missing event_id", http.StatusBadRequest)
		return
	}

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		http.Error(w, "Invalid event_id format", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	chat, err := h.chatService.GetOrCreateChat(r.Context(), eventID)
	if err != nil {
		http.Error(w, "Failed to get chat", http.StatusInternalServerError)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: false,
	})
	if err != nil {
		return
	}

	client := ws.NewClient(ws.ClientConfig{
		Hub:            h.hub,
		Conn:           conn,
		EventID:        eventID,
		UserID:         userID,
		PingInterval:   h.wsCfg.PingInterval,
		PongWait:       h.wsCfg.PongWait,
		WriteWait:      h.wsCfg.WriteWait,
		MaxMessageSize: h.wsCfg.MaxMessageSize,
		SendBuffer:     h.wsCfg.BroadcastBuffer,
	})

	h.hub.RegisterClient(client)

	ctx := r.Context()
	go client.WritePump(ctx)

	client.ReadPump(ctx, func(ctx context.Context, senderID uuid.UUID, incoming domain.IncomingMessage) error {
		username := r.Header.Get("X-User-Name")
		if username == "" {
			username = "anonymous"
		}

		message, err := h.chatService.CreateMessage(ctx, chat.ID, senderID, username, incoming)
		if err != nil {
			return err
		}

		outgoing := domain.OutgoingMessage{
			Message:   *message,
			EventType: domain.EventTypeNewMessage,
		}

		h.hub.Broadcast(eventID, outgoing)
		return nil
	})
}
