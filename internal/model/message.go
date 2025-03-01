package model

import (
	"time"

	"github.com/aidosgal/alem.core-service/internal/dto"
)

type Message struct {
	Id         int       `json:"id" db:"id"`
	SenderId   int       `json:"sender_id" db:"sender_id"`
	ReceiverId int       `json:"receiver_id" db:"receiver_id"`
	Text       *string   `json:"text" db:"text"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
	Files      []MessageFile `json:"files,omitempty"`
}

type MessageFile struct {
	Id        int       `json:"id" db:"id"`
	MessageId int       `json:"message_id" db:"message_id"`
	FileUrl   string    `json:"file_url" db:"file_url"`
	FileType  string    `json:"file_type" db:"file_type"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (m Message) ToDTO() dto.Message {
	files := make([]*dto.MessageFile, 0, len(m.Files))
	for _, f := range m.Files {
		fileDTO := f.ToDTO()
		files = append(files, &fileDTO)
	}

	return dto.Message{
		Id:         m.Id,
		SenderId:   m.SenderId,
		ReceiverId: m.ReceiverId,
		Text:       m.Text,
		Files:      files,
		CreatedAt:  m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  m.UpdatedAt.Format(time.RFC3339),
	}
}

func (f MessageFile) ToDTO() dto.MessageFile {
	return dto.MessageFile{
		Id:        f.Id,
		MessageId: f.MessageId,
		FileUrl:   f.FileUrl,
		FileType:  f.FileType,
	}
}

func MessageFromDTO(d dto.Message) Message {
	createdAt, _ := time.Parse(time.RFC3339, d.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, d.UpdatedAt)
	
	files := make([]MessageFile, 0, len(d.Files))
	for _, f := range d.Files {
		if f != nil {
			files = append(files, MessageFileFromDTO(*f))
		}
	}

	return Message{
		Id:         d.Id,
		SenderId:   d.SenderId,
		ReceiverId: d.ReceiverId,
		Text:       d.Text,
		Files:      files,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}
}

func MessageFileFromDTO(d dto.MessageFile) MessageFile {
	return MessageFile{
		Id:        d.Id,
		MessageId: d.MessageId,
		FileUrl:   d.FileUrl,
		FileType:  d.FileType,
		CreatedAt: time.Now(),
	}
}
