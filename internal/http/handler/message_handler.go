package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/aidosgal/alem.core-service/internal/dto"
	"github.com/aidosgal/alem.core-service/internal/http/middleware"
	"github.com/aidosgal/alem.core-service/internal/lib"
	"github.com/aidosgal/alem.core-service/internal/service"
	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	chatService service.ChatService
	upgrader    websocket.Upgrader
}

func NewWebSocketHandler(chatService service.ChatService) *WebSocketHandler {
	return &WebSocketHandler{
		chatService: chatService,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	userId64, ok := middleware.GetUserID(r)
	if !ok {
		lib.WriteError(w, http.StatusUnauthorized, fmt.Errorf("Unauthorized"))
	}

	userId := int(userId64)

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	h.chatService.RegisterConnection(userId, conn)
	log.Printf("User %d connected", userId)

	defer func() {
		conn.Close()
		h.chatService.RemoveConnection(userId)
		log.Printf("User %d disconnected", userId)
	}()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		if messageType == websocket.PingMessage {
			if err := conn.WriteMessage(websocket.PongMessage, nil); err != nil {
				log.Printf("Failed to send pong: %v", err)
				break
			}
			continue
		}

		var message dto.Message
		if err := json.Unmarshal(p, &message); err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}

		log.Printf("Received message from user %d to user %d", message.SenderId, message.ReceiverId)
	}
}

func (h *WebSocketHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	receiverIdStr := r.FormValue("receiver_id")
	text := r.FormValue("text")

	userId64, ok := middleware.GetUserID(r)
	if !ok {
		lib.WriteError(w, http.StatusUnauthorized, fmt.Errorf("Unauthorized"))
	}

	senderId := int(userId64)


	receiverId, err := strconv.Atoi(receiverIdStr)
	if err != nil {
		http.Error(w, "Invalid receiver ID", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	message, err := h.chatService.SendMessage(ctx, senderId, receiverId, text, files)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to send message: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(message)
}

func (h *WebSocketHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	receiverIdStr := r.URL.Query().Get("receiver_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

    userId64, ok := middleware.GetUserID(r)
	if !ok {
		lib.WriteError(w, http.StatusUnauthorized, fmt.Errorf("Unauthorized"))
	}

	senderId := int(userId64)

	receiverId, err := strconv.Atoi(receiverIdStr)
	if err != nil {
		http.Error(w, "Invalid receiver ID", http.StatusBadRequest)
		return
	}

	limit := 50 
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Get messages
	messages, err := h.chatService.GetMessagesByRoom(ctx, senderId, receiverId, limit, offset)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get messages: %v", err), http.StatusInternalServerError)
		return
	}

	// Return messages
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func (h *WebSocketHandler) GetRooms(w http.ResponseWriter, r *http.Request) {
	userId64, ok := middleware.GetUserID(r)
	if !ok {
		lib.WriteError(w, http.StatusUnauthorized, fmt.Errorf("Unauthorized"))
	}

	userId := int(userId64)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Get rooms
	messages, err := h.chatService.GetRoomsBySenderId(ctx, userId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get rooms: %v", err), http.StatusInternalServerError)
		return
	}

	// Return messages (representing the latest message from each room)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
