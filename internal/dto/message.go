package dto

type Message struct {
	Id         int            `json:"id"`
	SenderId   int            `json:"sender_id"`
	ReceiverId int            `json:"receiver_id"`
	Text       *string        `json:"text"`
	Files      []*MessageFile `json:"files"`
	CreatedAt  string         `json:"created_at"`
	UpdatedAt  string         `json:"updated_at"`
}

type MessageFile struct {
	Id        int    `json:"id"`
	MessageId int    `json:"message_id"`
	FileUrl   string `json:"file_url"`
	FileType  string `json:"file_type"`
}
