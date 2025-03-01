package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aidosgal/alem.core-service/internal/dto"
	"github.com/aidosgal/alem.core-service/internal/model"
	"github.com/aidosgal/alem.core-service/internal/repository"
	"github.com/gorilla/websocket"
)

type ChatService interface {
	SendMessage(ctx context.Context, senderId, receiverId int, text string, files []*multipart.FileHeader) (dto.Message, error)
	GetMessagesByRoom(ctx context.Context, senderId, receiverId, limit, offset int) ([]dto.Message, error)
	GetRoomsBySenderId(ctx context.Context, senderId int) ([]dto.Message, error)
	BroadcastMessage(ctx context.Context, message dto.Message) error
	RegisterConnection(userId int, conn *websocket.Conn)
	RemoveConnection(userId int)
}

type chatService struct {
	messageRepo      repository.MessageRepository
	connections      map[int]*websocket.Conn
	connectionsMutex sync.RWMutex
	publicDir        string
}

func NewChatService(messageRepo repository.MessageRepository, publicDir string) ChatService {
	return &chatService{
		messageRepo: messageRepo,
		connections: make(map[int]*websocket.Conn),
		publicDir:   publicDir,
	}
}

func (s *chatService) SendMessage(ctx context.Context, senderId, receiverId int, text string, files []*multipart.FileHeader) (dto.Message, error) {
	// Create message in database
	textPtr := &text
	newMessage := model.Message{
		SenderId:   senderId,
		ReceiverId: receiverId,
		Text:       textPtr,
	}

	savedMessage, err := s.messageRepo.CreateMessage(ctx, newMessage)
	if err != nil {
		return dto.Message{}, fmt.Errorf("failed to save message: %w", err)
	}

	// Process files if any
	if len(files) > 0 {
		savedMessage.Files = make([]model.MessageFile, 0, len(files))
		
		for _, fileHeader := range files {
			// Save file to disk
			fileUrl, fileType, err := s.saveFile(fileHeader, savedMessage.Id)
			if err != nil {
				return dto.Message{}, fmt.Errorf("failed to save file: %w", err)
			}

			// Create file record in database
			newFile := model.MessageFile{
				MessageId: savedMessage.Id,
				FileUrl:   fileUrl,
				FileType:  fileType,
			}

			savedFile, err := s.messageRepo.CreateMessageFile(ctx, newFile)
			if err != nil {
				return dto.Message{}, fmt.Errorf("failed to save file record: %w", err)
			}

			savedMessage.Files = append(savedMessage.Files, savedFile)
		}
	}

	// Convert to DTO
	messageDTO := savedMessage.ToDTO()

	// Broadcast message to connected clients
	go s.BroadcastMessage(context.Background(), messageDTO)

	return messageDTO, nil
}

func (s *chatService) saveFile(fileHeader *multipart.FileHeader, messageId int) (string, string, error) {
	// Create directory if it doesn't exist
	fileDir := filepath.Join(s.publicDir, "files", fmt.Sprintf("message_%d", messageId))
	if err := os.MkdirAll(fileDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Get file extension
	ext := filepath.Ext(fileHeader.Filename)
	fileType := strings.TrimPrefix(ext, ".")
	if fileType == "" {
		fileType = "unknown"
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), strings.ReplaceAll(fileHeader.Filename, " ", "_"), ext)
	filePath := filepath.Join(fileDir, filename)

	// Open source file
	src, err := fileHeader.Open()
	if err != nil {
		return "", "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy file content
	if _, err = io.Copy(dst, src); err != nil {
		return "", "", fmt.Errorf("failed to copy file content: %w", err)
	}

	// Return relative URL for accessing the file
	fileUrl := fmt.Sprintf("/files/message_%d/%s", messageId, filename)
	return fileUrl, fileType, nil
}

func (s *chatService) GetMessagesByRoom(ctx context.Context, senderId, receiverId, limit, offset int) ([]dto.Message, error) {
	messages, err := s.messageRepo.GetMessagesByRoom(ctx, senderId, receiverId, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	// Convert to DTOs
	messageDTOs := make([]dto.Message, 0, len(messages))
	for _, message := range messages {
		messageDTOs = append(messageDTOs, message.ToDTO())
	}

	return messageDTOs, nil
}

func (s *chatService) GetRoomsBySenderId(ctx context.Context, senderId int) ([]dto.Message, error) {
	messages, err := s.messageRepo.GetRoomsBySenderId(ctx, senderId)
	if err != nil {
		return nil, fmt.Errorf("failed to get rooms: %w", err)
	}

	// Convert to DTOs
	messageDTOs := make([]dto.Message, 0, len(messages))
	for _, message := range messages {
		messageDTOs = append(messageDTOs, message.ToDTO())
	}

	return messageDTOs, nil
}

func (s *chatService) BroadcastMessage(ctx context.Context, message dto.Message) error {
	s.connectionsMutex.RLock()
	defer s.connectionsMutex.RUnlock()

	// Try to send to both sender and receiver (if connected)
	sendErr := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(2)

	// Function to send message to a specific user
	sendToUser := func(userId int) {
		defer wg.Done()
		conn, ok := s.connections[userId]
		if !ok {
			return // User not connected
		}

		if err := conn.WriteJSON(message); err != nil {
			select {
			case sendErr <- fmt.Errorf("failed to send message to user %d: %w", userId, err):
			default:
			}
		}
	}

	// Send to both users
	go sendToUser(message.SenderId)
	go sendToUser(message.ReceiverId)

	// Wait for both sends to complete
	wg.Wait()
	
	// Check if any errors occurred
	select {
	case err := <-sendErr:
		return err
	default:
		return nil
	}
}

func (s *chatService) RegisterConnection(userId int, conn *websocket.Conn) {
	s.connectionsMutex.Lock()
	defer s.connectionsMutex.Unlock()
	
	// Close existing connection if any
	if existingConn, ok := s.connections[userId]; ok {
		existingConn.Close()
	}
	
	s.connections[userId] = conn
}

func (s *chatService) RemoveConnection(userId int) {
	s.connectionsMutex.Lock()
	defer s.connectionsMutex.Unlock()
	
	delete(s.connections, userId)
}
