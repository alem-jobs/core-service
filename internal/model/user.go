package model

type User struct {
	Id             int     `json:"id"`
	Name           string  `json:"name"`
	OrganizationId int     `json:"organization_id"`
	Phone         string  `json:"phone"`
	Password      string  `json:"password"`
	AvatarURL     string  `json:"avatar_url"`
	Balance       float64 `json:"balance"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}
