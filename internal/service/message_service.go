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
	userService      *UserService
}

func NewChatService(messageRepo repository.MessageRepository, publicDir string, userService *UserService) ChatService {
	return &chatService{
		messageRepo: messageRepo,
		connections: make(map[int]*websocket.Conn),
		publicDir:   publicDir,
		userService: userService,
	}
}

func (s *chatService) SendMessage(ctx context.Context, senderId, receiverId int, text string, files []*multipart.FileHeader) (dto.Message, error) {
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

	if len(files) > 0 {
		savedMessage.Files = make([]model.MessageFile, 0, len(files))

		for _, fileHeader := range files {
			fileUrl, fileType, err := s.saveFile(fileHeader, savedMessage.Id)
			if err != nil {
				return dto.Message{}, fmt.Errorf("failed to save file: %w", err)
			}

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

	messageDTO := savedMessage.ToDTO()

	go s.BroadcastMessage(context.Background(), messageDTO)

	return messageDTO, nil
}

func (s *chatService) saveFile(fileHeader *multipart.FileHeader, messageId int) (string, string, error) {
	fileDir := filepath.Join(s.publicDir, "files", fmt.Sprintf("message_%d", messageId))
	if err := os.MkdirAll(fileDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create directory: %w", err)
	}

	ext := filepath.Ext(fileHeader.Filename)
	fileType := strings.TrimPrefix(ext, ".")
	if fileType == "" {
		fileType = "unknown"
	}

	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), strings.ReplaceAll(fileHeader.Filename, " ", "_"), ext)
	filePath := filepath.Join(fileDir, filename)

	src, err := fileHeader.Open()
	if err != nil {
		return "", "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return "", "", fmt.Errorf("failed to copy file content: %w", err)
	}

	fileUrl := fmt.Sprintf("/files/message_%d/%s", messageId, filename)
	return fileUrl, fileType, nil
}

func (s *chatService) GetMessagesByRoom(ctx context.Context, senderId, receiverId, limit, offset int) ([]dto.Message, error) {
	messages, err := s.messageRepo.GetMessagesByRoom(ctx, senderId, receiverId, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	messageDTOs := make([]dto.Message, 0, len(messages))
	for _, message := range messages {
		receiver, err := s.userService.GetProfile(receiverId)
		if err != nil {
			return nil, err
		}

		messageDTO := message.ToDTO()
		messageDTO.Receiver = receiver

		messageDTOs = append(messageDTOs, messageDTO)
	}

	return messageDTOs, nil
}

func (s *chatService) GetRoomsBySenderId(ctx context.Context, senderId int) ([]dto.Message, error) {
	messages, err := s.messageRepo.GetRoomsBySenderId(ctx, senderId)
	if err != nil {
		return nil, fmt.Errorf("failed to get rooms: %w", err)
	}

	messageDTOs := make([]dto.Message, 0, len(messages))
	for _, message := range messages {
		receiver, err := s.userService.GetProfile(message.ReceiverId)
		if err != nil {
			return nil, err
		}

		messageDTO := message.ToDTO()
		messageDTO.Receiver = receiver

		messageDTOs = append(messageDTOs, messageDTO)
	}

	return messageDTOs, nil
}

func (s *chatService) BroadcastMessage(ctx context.Context, message dto.Message) error {
	s.connectionsMutex.RLock()
	defer s.connectionsMutex.RUnlock()

	sendErr := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(2)

	sendToUser := func(userId int) {
		defer wg.Done()
		conn, ok := s.connections[userId]
		if !ok {
			return
		}

		if err := conn.WriteJSON(message); err != nil {
			select {
			case sendErr <- fmt.Errorf("failed to send message to user %d: %w", userId, err):
			default:
			}
		}
	}

	go sendToUser(message.SenderId)
	go sendToUser(message.ReceiverId)

	wg.Wait()

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
