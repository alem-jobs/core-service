package dto

type LoginRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type LoginResponse struct {
	User        User   `json:"user"`
	IsCompleted bool   `json:"is_completed"`
	Token       string `json:"token"`
}

type RegisterRequest struct {
	User User `json:"user"`
}

type RegisterResponse struct {
	User  User   `json:"user"`
	Token string `json:"token"`
}

type User struct {
	Id             int     `json:"id"`
	Name           string  `json:"name"`
	OrganizationId int     `json:"organization_id"`
	Phone          string  `json:"phone"`
	Password       string  `json:"password"`
	AvatarURL      string  `json:"avatar_url"`
	Balance        float64 `json:"balance"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}
