package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/aidosgal/alem.core-service/internal/model"
)

type MessageRepository interface {
	CreateMessage(ctx context.Context, message model.Message) (model.Message, error)
	CreateMessageFile(ctx context.Context, file model.MessageFile) (model.MessageFile, error)
	GetMessagesByRoom(ctx context.Context, senderId, receiverId int, limit, offset int) ([]model.Message, error)
	GetRoomsBySenderId(ctx context.Context, senderId int) ([]model.Message, error)
	GetMessageById(ctx context.Context, id int) (model.Message, error)
}

type messageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) MessageRepository {
	return &messageRepository{
		db: db,
	}
}

func (r *messageRepository) CreateMessage(ctx context.Context, message model.Message) (model.Message, error) {
	query := `
		INSERT INTO messages (sender_id, receiver_id, text, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, sender_id, receiver_id, text, created_at, updated_at
	`

	now := time.Now()
	message.CreatedAt = now
	message.UpdatedAt = now

	var result model.Message
	row := r.db.QueryRowContext(
		ctx,
		query,
		message.SenderId,
		message.ReceiverId,
		message.Text,
		message.CreatedAt,
		message.UpdatedAt,
	)

	err := row.Scan(
		&result.Id,
		&result.SenderId,
		&result.ReceiverId,
		&result.Text,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		return model.Message{}, fmt.Errorf("failed to create message: %w", err)
	}

	return result, nil
}

func (r *messageRepository) CreateMessageFile(ctx context.Context, file model.MessageFile) (model.MessageFile, error) {
	query := `
		INSERT INTO message_files (message_id, file_url, file_type, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, message_id, file_url, file_type, created_at
	`

	file.CreatedAt = time.Now()

	var result model.MessageFile
	row := r.db.QueryRowContext(
		ctx,
		query,
		file.MessageId,
		file.FileUrl,
		file.FileType,
		file.CreatedAt,
	)

	err := row.Scan(
		&result.Id,
		&result.MessageId,
		&result.FileUrl,
		&result.FileType,
		&result.CreatedAt,
	)

	if err != nil {
		return model.MessageFile{}, fmt.Errorf("failed to create message file: %w", err)
	}

	return result, nil
}

func (r *messageRepository) GetMessagesByRoom(ctx context.Context, senderId, receiverId int, limit, offset int) ([]model.Message, error) {
	query := `
		SELECT id, sender_id, receiver_id, text, created_at, updated_at
		FROM messages
		WHERE (sender_id = $1 AND receiver_id = $2) OR (sender_id = $2 AND receiver_id = $1)
		ORDER BY created_at ASC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, senderId, receiverId, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		var message model.Message
		err := rows.Scan(
			&message.Id,
			&message.SenderId,
			&message.ReceiverId,
			&message.Text,
			&message.CreatedAt,
			&message.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message row: %w", err)
		}

		// Fetch files for each message
		files, err := r.getMessageFiles(ctx, message.Id)
		if err != nil {
			return nil, err
		}
		message.Files = files

		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating message rows: %w", err)
	}

	return messages, nil
}

func (r *messageRepository) getMessageFiles(ctx context.Context, messageId int) ([]model.MessageFile, error) {
	query := `
		SELECT id, message_id, file_url, file_type, created_at
		FROM message_files
		WHERE message_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, messageId)
	if err != nil {
		return nil, fmt.Errorf("failed to get message files: %w", err)
	}
	defer rows.Close()

	var files []model.MessageFile
	for rows.Next() {
		var file model.MessageFile
		err := rows.Scan(
			&file.Id,
			&file.MessageId,
			&file.FileUrl,
			&file.FileType,
			&file.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message file row: %w", err)
		}
		files = append(files, file)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating message file rows: %w", err)
	}

	return files, nil
}

func (r *messageRepository) GetRoomsBySenderId(ctx context.Context, senderId int) ([]model.Message, error) {
	query := `
		SELECT DISTINCT ON (room_identifier) id, sender_id, receiver_id, text, created_at, updated_at,
		CASE 
			WHEN sender_id < receiver_id THEN CONCAT(sender_id, '-', receiver_id)
			ELSE CONCAT(receiver_id, '-', sender_id)
		END as room_identifier
		FROM messages
		WHERE sender_id = $1 OR receiver_id = $1
		ORDER BY room_identifier, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, senderId)
	if err != nil {
		return nil, fmt.Errorf("failed to get rooms: %w", err)
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		var message model.Message
		var roomIdentifier string
		err := rows.Scan(
			&message.Id,
			&message.SenderId,
			&message.ReceiverId,
			&message.Text,
			&message.CreatedAt,
			&message.UpdatedAt,
			&roomIdentifier, // We're scanning this but not using it directly in the model
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan room row: %w", err)
		}

		// Fetch files for each message
		files, err := r.getMessageFiles(ctx, message.Id)
		if err != nil {
			return nil, err
		}
		message.Files = files

		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating room rows: %w", err)
	}

	return messages, nil
}

func (r *messageRepository) GetMessageById(ctx context.Context, id int) (model.Message, error) {
	query := `
		SELECT id, sender_id, receiver_id, text, created_at, updated_at
		FROM messages
		WHERE id = $1
	`

	var message model.Message
	row := r.db.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&message.Id,
		&message.SenderId,
		&message.ReceiverId,
		&message.Text,
		&message.CreatedAt,
		&message.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Message{}, fmt.Errorf("message not found: %w", err)
		}
		return model.Message{}, fmt.Errorf("failed to get message: %w", err)
	}

	// Fetch files
	files, err := r.getMessageFiles(ctx, message.Id)
	if err != nil {
		return model.Message{}, err
	}
	message.Files = files

	return message, nil
}
