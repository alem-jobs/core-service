package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
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
	channels    map[string]*Channel
	mutex       sync.RWMutex
}

type Channel struct {
	connections map[int]*websocket.Conn
	mutex       sync.Mutex
}

func generateChannelID(user1, user2 int) string {
	// Make sure the smaller ID is always first for consistency
	if user1 > user2 {
		user1, user2 = user2, user1
	}
	return fmt.Sprintf("%d-%d", user1, user2)
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
		channels: make(map[string]*Channel),
	}
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from request
	userIDstr := r.URL.Query().Get("sender_id")
	userID, _ := strconv.Atoi(userIDstr)

	// Extract receiver ID from query parameters
	receiverIDStr := r.URL.Query().Get("receiver_id")
	if receiverIDStr == "" {
		http.Error(w, "Receiver ID is required", http.StatusBadRequest)
		return
	}

	receiverID, err := strconv.Atoi(receiverIDStr)
	if err != nil {
		http.Error(w, "Invalid receiver ID", http.StatusBadRequest)
		return
	}

	// Generate a unique channel ID for these two users
	channelID := generateChannelID(userID, receiverID)

	// Upgrade HTTP connection to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// Add this connection to the appropriate channel
	h.addToChannel(channelID, userID, conn)
	log.Printf("User %d connected to channel with user %d (Channel ID: %s)", userID, receiverID, channelID)

	// Clean up when the connection closes
	defer func() {
		conn.Close()
		h.removeFromChannel(channelID, userID)
		log.Printf("User %d disconnected from channel with user %d", userID, receiverID)
	}()

	// Message handling loop
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle ping messages
		if messageType == websocket.PingMessage {
			if err := conn.WriteMessage(websocket.PongMessage, nil); err != nil {
				log.Printf("Failed to send pong: %v", err)
				break
			}
			continue
		}

		// Parse the incoming message
		var message dto.Message
		if err := json.Unmarshal(p, &message); err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}

		// Validate sender and receiver
		if message.SenderId != userID || message.ReceiverId != receiverID {
			log.Printf("Invalid sender or receiver ID in message")
			continue
		}

		if err != nil {
			log.Printf("Failed to save message: %v", err)
			continue
		}

		log.Printf("Message processed from user %d to user %d", message.SenderId, message.ReceiverId)
	}
}

// Add a connection to a specific channel
func (h *WebSocketHandler) addToChannel(channelID string, userID int, conn *websocket.Conn) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Create the channel if it doesn't exist
	if _, exists := h.channels[channelID]; !exists {
		h.channels[channelID] = &Channel{
			connections: make(map[int]*websocket.Conn),
		}
	}

	// Add the connection to the channel
	channel := h.channels[channelID]
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	channel.connections[userID] = conn
}

// Remove a connection from a channel
func (h *WebSocketHandler) removeFromChannel(channelID string, userID int) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Check if the channel exists
	channel, exists := h.channels[channelID]
	if !exists {
		return
	}

	// Remove the connection
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	delete(channel.connections, userID)

	// Clean up empty channels
	if len(channel.connections) == 0 {
		delete(h.channels, channelID)
	}
}

// Broadcast a message to all connections in a channel
func (h *WebSocketHandler) broadcastToChannel(channelID string, message []byte) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	// Check if the channel exists
	channel, exists := h.channels[channelID]
	if !exists {
		return
	}

	// Send the message to all connections in the channel
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	for userID, conn := range channel.connections {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Failed to send message to user %d: %v", userID, err)
		}
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
		return
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

	// Also broadcast the message through WebSocket for real-time updates
	channelID := generateChannelID(senderId, receiverId)
	messageJSON, _ := json.Marshal(message)
	h.broadcastToChannel(channelID, messageJSON)

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
		return
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
		return
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
