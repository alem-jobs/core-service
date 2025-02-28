package dto

import (
	"github.com/aidosgal/alem.core-service/internal/model"
)

type LoginRequest struct {
	Phone string `json:"phone"`
	Password string `json:"password"`
}

type LoginResponse struct {
	User model.User `json:"user"`
}

type RegisterRequest struct {
	User model.User `json:"user"`
}

type RegisterResponse struct {
	User model.User `json:"user"`
}
