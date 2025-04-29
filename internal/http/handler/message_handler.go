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
				log.Printf("CheckOrigin called for request from: %s", r.RemoteAddr)
				return true
			},
		},
		channels: make(map[string]*Channel),
	}
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Printf("WebSocket connection attempt from %s", r.RemoteAddr)
	log.Printf("Request URL: %s", r.URL.String())
	log.Printf("Request Headers: %v", r.Header)

	// Extract user ID from request
	userIDstr := r.URL.Query().Get("sender_id")
	log.Printf("Extracted sender_id from query: %s", userIDstr)

	if userIDstr == "" {
		log.Printf("Error: Missing sender_id parameter")
		http.Error(w, "Sender ID is required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDstr)
	if err != nil {
		log.Printf("Error: Invalid sender_id format: %v", err)
		http.Error(w, "Invalid sender ID", http.StatusBadRequest)
		return
	}

	// Extract receiver ID from query parameters
	receiverIDStr := r.URL.Query().Get("receiver_id")
	log.Printf("Extracted receiver_id from query: %s", receiverIDStr)

	if receiverIDStr == "" {
		log.Printf("Error: Missing receiver_id parameter")
		http.Error(w, "Receiver ID is required", http.StatusBadRequest)
		return
	}

	receiverID, err := strconv.Atoi(receiverIDStr)
	if err != nil {
		log.Printf("Error: Invalid receiver_id format: %v", err)
		http.Error(w, "Invalid receiver ID", http.StatusBadRequest)
		return
	}

	// Generate a unique channel ID for these two users
	channelID := generateChannelID(userID, receiverID)
	log.Printf("Generated channel ID: %s for users %d and %d", channelID, userID, receiverID)

	// Upgrade HTTP connection to WebSocket
	log.Printf("Attempting to upgrade connection to WebSocket")
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	log.Printf("Successfully upgraded connection to WebSocket")

	// Add this connection to the appropriate channel
	h.addToChannel(channelID, userID, conn)
	log.Printf("User %d connected to channel with user %d (Channel ID: %s)", userID, receiverID, channelID)

	defer func() {
		log.Printf("Closing WebSocket connection for user %d in channel %s", userID, channelID)
		conn.Close()
		h.removeFromChannel(channelID, userID)
		log.Printf("User %d disconnected from channel with user %d", userID, receiverID)
	}()

	for {
		log.Printf("Waiting for message from user %d", userID)
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			} else {
				log.Printf("Connection closed: %v", err)
			}
			break
		}
		log.Printf("Received message of type %d from user %d", messageType, userID)

		if messageType == websocket.PingMessage {
			log.Printf("Received ping from user %d, sending pong", userID)
			if err := conn.WriteMessage(websocket.PongMessage, nil); err != nil {
				log.Printf("Failed to send pong: %v", err)
				break
			}
			continue
		}

		log.Printf("Parsing message: %s", string(p))
		var messageRequest dto.Message
		if err := json.Unmarshal(p, &messageRequest); err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}
		log.Printf("Successfully parsed message from user %d to user %d", messageRequest.SenderId, messageRequest.ReceiverId)

		if messageRequest.SenderId != userID || messageRequest.ReceiverId != receiverID {
			log.Printf("Invalid sender or receiver ID in message. Expected sender: %d, got: %d. Expected receiver: %d, got: %d",
				userID, messageRequest.SenderId, receiverID, messageRequest.ReceiverId)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		message, err := h.chatService.SendMessage(ctx, userID, receiverID, *messageRequest.Text, nil)
		cancel()

		if err != nil {
			log.Printf("Failed to save message: %v", err)
			if err := conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "Failed to save message"}`)); err != nil {
				log.Printf("Failed to send error response: %v", err)
			}
			continue
		}

		// Broadcast the message to all connections in the channel
		messageJSON, err := json.Marshal(message)
		if err != nil {
			log.Printf("Failed to marshal message: %v", err)
			continue
		}

		h.broadcastToChannel(channelID, messageJSON)
		log.Printf("Message processed and broadcast from user %d to user %d", message.SenderId, message.ReceiverId)
	}
}

// Add a connection to a specific channel
func (h *WebSocketHandler) addToChannel(channelID string, userID int, conn *websocket.Conn) {
	log.Printf("Adding user %d to channel %s", userID, channelID)
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Create the channel if it doesn't exist
	if _, exists := h.channels[channelID]; !exists {
		log.Printf("Creating new channel %s", channelID)
		h.channels[channelID] = &Channel{
			connections: make(map[int]*websocket.Conn),
		}
	}

	// Add the connection to the channel
	channel := h.channels[channelID]
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	channel.connections[userID] = conn
	log.Printf("Added user %d to channel %s. Total connections in channel: %d", userID, channelID, len(channel.connections))
}

// Remove a connection from a channel
func (h *WebSocketHandler) removeFromChannel(channelID string, userID int) {
	log.Printf("Removing user %d from channel %s", userID, channelID)
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Check if the channel exists
	channel, exists := h.channels[channelID]
	if !exists {
		log.Printf("Channel %s does not exist", channelID)
		return
	}

	// Remove the connection
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	delete(channel.connections, userID)
	log.Printf("Removed user %d from channel %s. Remaining connections: %d", userID, channelID, len(channel.connections))

	// Clean up empty channels
	if len(channel.connections) == 0 {
		log.Printf("Channel %s is empty, removing it", channelID)
		delete(h.channels, channelID)
	}
}

// Broadcast a message to all connections in a channel
func (h *WebSocketHandler) broadcastToChannel(channelID string, message []byte) {
	log.Printf("Broadcasting message to channel %s", channelID)
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	// Check if the channel exists
	channel, exists := h.channels[channelID]
	if !exists {
		log.Printf("Channel %s does not exist, cannot broadcast", channelID)
		return
	}

	// Send the message to all connections in the channel
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	log.Printf("Sending message to %d connections in channel %s", len(channel.connections), channelID)
	for userID, conn := range channel.connections {
		log.Printf("Sending message to user %d in channel %s", userID, channelID)
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Failed to send message to user %d: %v", userID, err)
		} else {
			log.Printf("Successfully sent message to user %d", userID)
		}
	}
}

func (h *WebSocketHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	log.Printf("SendMessage called from %s", r.RemoteAddr)
	log.Printf("Request method: %s, content type: %s", r.Method, r.Header.Get("Content-Type"))

	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		log.Printf("Failed to parse form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	log.Printf("Form parsed successfully")

	// Log all form values for debugging
	log.Printf("Form values: %v", r.Form)
	log.Printf("PostForm values: %v", r.PostForm)

	receiverIdStr := r.FormValue("receiver_id")
	text := r.FormValue("text")
	log.Printf("Receiver ID: %s, Text: %s", receiverIdStr, text)

	userId64, ok := middleware.GetUserID(r)
	if !ok {
		log.Printf("Failed to get user ID from request")
		lib.WriteError(w, http.StatusUnauthorized, fmt.Errorf("Unauthorized"))
		return
	}
	log.Printf("User ID from middleware: %d", userId64)

	senderId := int(userId64)

	receiverId, err := strconv.Atoi(receiverIdStr)
	if err != nil {
		log.Printf("Invalid receiver ID: %v", err)
		http.Error(w, "Invalid receiver ID", http.StatusBadRequest)
		return
	}
	log.Printf("Parsed receiver ID: %d", receiverId)

	files := r.MultipartForm.File["files"]
	log.Printf("Number of files: %d", len(files))

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	log.Printf("Calling chatService.SendMessage with sender: %d, receiver: %d", senderId, receiverId)
	message, err := h.chatService.SendMessage(ctx, senderId, receiverId, text, files)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
		http.Error(w, fmt.Sprintf("Failed to send message: %v", err), http.StatusInternalServerError)
		return
	}
	log.Printf("Message sent successfully: %+v", message)

	// Also broadcast the message through WebSocket for real-time updates
	channelID := generateChannelID(senderId, receiverId)
	messageJSON, _ := json.Marshal(message)
	log.Printf("Broadcasting message to channel %s", channelID)
	h.broadcastToChannel(channelID, messageJSON)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(message)
	log.Printf("Response sent successfully")
}

func (h *WebSocketHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	log.Printf("GetMessages called from %s", r.RemoteAddr)
	log.Printf("Query parameters: %v", r.URL.Query())

	receiverIdStr := r.URL.Query().Get("receiver_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	log.Printf("Receiver ID: %s, Limit: %s, Offset: %s", receiverIdStr, limitStr, offsetStr)

	userId64, ok := middleware.GetUserID(r)
	if !ok {
		log.Printf("Failed to get user ID from request")
		lib.WriteError(w, http.StatusUnauthorized, fmt.Errorf("Unauthorized"))
		return
	}
	log.Printf("User ID from middleware: %d", userId64)

	senderId := int(userId64)

	if receiverIdStr == "" {
		log.Printf("Missing receiver_id parameter")
		http.Error(w, "Receiver ID is required", http.StatusBadRequest)
		return
	}

	receiverId, err := strconv.Atoi(receiverIdStr)
	if err != nil {
		log.Printf("Invalid receiver ID: %v", err)
		http.Error(w, "Invalid receiver ID", http.StatusBadRequest)
		return
	}
	log.Printf("Parsed receiver ID: %d", receiverId)

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		} else if err != nil {
			log.Printf("Invalid limit parameter: %v", err)
		}
	}
	log.Printf("Using limit: %d", limit)

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		} else if err != nil {
			log.Printf("Invalid offset parameter: %v", err)
		}
	}
	log.Printf("Using offset: %d", offset)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Get messages
	log.Printf("Calling chatService.GetMessagesByRoom with sender: %d, receiver: %d, limit: %d, offset: %d",
		senderId, receiverId, limit, offset)
	messages, err := h.chatService.GetMessagesByRoom(ctx, senderId, receiverId, limit, offset)
	if err != nil {
		log.Printf("Failed to get messages: %v", err)
		http.Error(w, fmt.Sprintf("Failed to get messages: %v", err), http.StatusInternalServerError)
		return
	}
	log.Printf("Retrieved %d messages successfully", len(messages))

	// Return messages
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
	log.Printf("Response sent successfully")
}

func (h *WebSocketHandler) GetRooms(w http.ResponseWriter, r *http.Request) {
	log.Printf("GetRooms called from %s", r.RemoteAddr)

	userId64, ok := middleware.GetUserID(r)
	if !ok {
		log.Printf("Failed to get user ID from request")
		lib.WriteError(w, http.StatusUnauthorized, fmt.Errorf("Unauthorized"))
		return
	}
	log.Printf("User ID from middleware: %d", userId64)

	userId := int(userId64)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Get rooms
	log.Printf("Calling chatService.GetRoomsBySenderId with user ID: %d", userId)
	messages, err := h.chatService.GetRoomsBySenderId(ctx, userId)
	if err != nil {
		log.Printf("Failed to get rooms: %v", err)
		http.Error(w, fmt.Sprintf("Failed to get rooms: %v", err), http.StatusInternalServerError)
		return
	}
	log.Printf("Retrieved %d rooms successfully", len(messages))

	// Return messages (representing the latest message from each room)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
	log.Printf("Response sent successfully")
}
