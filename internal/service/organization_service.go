package service

import (
	"errors"
	"log/slog"

	"github.com/aidosgal/alem.core-service/internal/dto"
	"github.com/aidosgal/alem.core-service/internal/model"
	"github.com/aidosgal/alem.core-service/internal/repository"
)

type OrganizationService struct {
	log  *slog.Logger
	repo *repository.OrganizationRepository
	user *UserService
}

func NewOrganizationService(log *slog.Logger, repo *repository.OrganizationRepository, user *UserService) *OrganizationService {
	return &OrganizationService{
		log:  log,
		repo: repo,
		user: user,
	}
}

func (s *OrganizationService) CreateOrganization(req *dto.Organization) (*dto.Organization, error) {
	if req.Name == "" || req.Description == "" {
		s.log.Warn("Invalid organization data")
		return nil, errors.New("name and description cannot be empty")
	}

	org := &model.Organization{
		Name:        req.Name,
		Description: req.Description,
	}

	err := s.repo.CreateOrganization(org)
	if err != nil {
		s.log.Error("Failed to create organization", slog.Any("error", err))
		return nil, err
	}

	s.log.Info("Organization created successfully", slog.Int("id", org.Id))
	req.Id = org.Id
	return req, nil
}

func (s *OrganizationService) GetOrganization(id int) (*dto.Organization, error) {
	org, err := s.repo.GetOrganization(id)
	if err != nil {
		s.log.Error("Failed to retrieve organization", slog.Any("error", err))
		return nil, err
	}
	if org == nil {
		s.log.Warn("Organization not found", slog.Int("id", id))
		return nil, nil
	}
	users, err := s.user.ListUsers(org.Id)
	if err != nil {
		return nil, err
	}

	s.log.Info("Organization retrieved successfully", slog.Int("id", org.Id))
	return &dto.Organization{
		Id:          org.Id,
		Name:        org.Name,
		Description: org.Description,
		Users:       users,
	}, nil
}

func (s *OrganizationService) GetAllOrganizations() ([]*dto.Organization, error) {
	orgs, err := s.repo.GetAllOrganizations()
	if err != nil {
		s.log.Error("Failed to retrieve organizations", slog.Any("error", err))
		return nil, err
	}

	var dtoOrgs []*dto.Organization
	for _, org := range orgs {
		dtoOrgs = append(dtoOrgs, &dto.Organization{
			Id:          org.Id,
			Name:        org.Name,
			Description: org.Description,
		})
	}

	s.log.Info("Organizations retrieved successfully", slog.Int("count", len(dtoOrgs)))
	return dtoOrgs, nil
}
