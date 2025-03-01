package service

import (
	"errors"
	"log/slog"

	"github.com/aidosgal/alem.core-service/internal/dto"
	"github.com/aidosgal/alem.core-service/internal/lib"
	"github.com/aidosgal/alem.core-service/internal/model"
	"github.com/aidosgal/alem.core-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	log      *slog.Logger
	userRepo *repository.UserRepository
}

func NewUserService(log *slog.Logger, userRepo *repository.UserRepository) *UserService {
	return &UserService{
		log:      log,
		userRepo: userRepo,
	}
}

func (s *UserService) Register(req dto.RegisterRequest) (*dto.RegisterResponse, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.User.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	req.User.Password = string(hashedPassword)

	userModel := model.User{
		Name:           req.User.Name,
		OrganizationId: req.User.OrganizationId,
		Phone:          req.User.Phone,
		Password:       req.User.Password,
		AvatarURL:      req.User.AvatarURL,
		Balance:        0.0,
	}

	id, err := s.userRepo.CreateUser(&userModel)
	if err != nil {
		return nil, err
	}

	token, err := lib.NewToken(
		int64(id), int64(req.User.OrganizationId), "organization_type")
	if err != nil {
		return nil, err
	}

	return &dto.RegisterResponse{User: req.User, Token: token}, nil
}

func (s *UserService) Login(req dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.userRepo.GetUserByPhone(req.Phone)
	if err != nil {
		return nil, errors.New("invalid phone or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid phone or password")
	}

	token, err := lib.NewToken(int64(user.Id), int64(user.OrganizationId), "organization_type")
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		User: dto.User{
			Id:             user.Id,
			Name:           user.Name,
			OrganizationId: user.OrganizationId,
			Phone:          user.Phone,
			AvatarURL:      user.AvatarURL,
			Balance:        user.Balance,
		},
		IsCompleted: true,
		Token:       token,
	}, nil
}

func (s *UserService) GetProfile(userID int) (*dto.User, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	return &dto.User{
		Id:             user.Id,
		Name:           user.Name,
		OrganizationId: user.OrganizationId,
		Phone:          user.Phone,
		AvatarURL:      user.AvatarURL,
		Balance:        user.Balance,
	}, nil
}

func (s *UserService) ListUsers() ([]dto.User, error) {
	users, err := s.userRepo.ListUsers()
	if err != nil {
		return nil, err
	}

	var userDTOs []dto.User
	for _, user := range users {
		userDTOs = append(userDTOs, dto.User{
			Id:             user.Id,
			Name:           user.Name,
			OrganizationId: user.OrganizationId,
			Phone:          user.Phone,
			AvatarURL:      user.AvatarURL,
			Balance:        user.Balance,
		})
	}
	return userDTOs, nil
}
