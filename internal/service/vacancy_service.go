package service

import (
	"context"
	"log/slog"

	"github.com/aidosgal/alem.core-service/internal/dto"
	"github.com/aidosgal/alem.core-service/internal/model"
	"github.com/aidosgal/alem.core-service/internal/repository"
)

type VacancyService struct {
	log          *slog.Logger
	vacancy      *repository.VacancyRepository
	detail       *repository.VacancyDetailRepository
	organization *OrganizationService
}

func NewVacancyService(log *slog.Logger, vacancy *repository.VacancyRepository, detail *repository.VacancyDetailRepository, organization *OrganizationService) *VacancyService {
	return &VacancyService{
		log:          log,
		vacancy:      vacancy,
		detail:       detail,
		organization: organization,
	}
}

func (s *VacancyService) CreateVacancy(ctx context.Context, req dto.CreateVacancyRequest) (*dto.CreateVacancyResponse, error) {
	id, err := s.vacancy.Create(ctx, &model.Vacancy{
		Title:          req.Vacancy.Title,
		Description:    req.Vacancy.Description,
		SalaryFrom:     req.Vacancy.SalaryFrom,
		SalaryTo:       req.Vacancy.SalaryTo,
		SalaryExact:    req.Vacancy.SalaryExact,
		SalaryType:     req.Vacancy.SalaryType,
		SalaryCurrency: req.Vacancy.SalaryCurrency,
		OrganizationID: req.Vacancy.OrganizationID,
		CategoryID:     req.Vacancy.CategoryID,
		Country:        req.Vacancy.Country,
	})
	if err != nil {
		return nil, err
	}

	req.Vacancy.ID = id

	// Create details
	var details []dto.VacancyDetailResponse
	for _, detail := range req.Vacancy.Details {
		detailID, err := s.detail.Create(ctx, &model.VacancyDetail{
			GroupName: detail.GroupName,
			Name:      detail.Name,
			Value:     detail.Value,
			IconURL:   &detail.IconURL,
			VacancyID: id,
		})
		if err != nil {
			return nil, err
		}
		detail.ID = detailID
		details = append(details, detail)
	}

	req.Vacancy.Details = details
	return &dto.CreateVacancyResponse{Vacancy: req.Vacancy}, nil
}

func (s *VacancyService) GetVacancyByID(ctx context.Context, id int64) (*dto.Vacancy, error) {
	vacancy, err := s.vacancy.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	details, err := s.detail.GetByVacancyID(ctx, id)
	if err != nil {
		return nil, err
	}
	organization, err := s.organization.GetOrganization(int(vacancy.OrganizationID))
	if err != nil {
		return nil, err
	}

	var detailResponses []dto.VacancyDetailResponse
	for _, d := range details {
		detailResponses = append(detailResponses, dto.VacancyDetailResponse{
			ID:        d.ID,
			GroupName: d.GroupName,
			Name:      d.Name,
			Value:     d.Value,
			IconURL:   *d.IconURL,
			VacancyID: d.VacancyID,
		})
	}

	return &dto.Vacancy{
		ID:             vacancy.ID,
		Title:          vacancy.Title,
		Description:    vacancy.Description,
		SalaryFrom:     vacancy.SalaryFrom,
		SalaryTo:       vacancy.SalaryTo,
		SalaryExact:    vacancy.SalaryExact,
		SalaryType:     vacancy.SalaryType,
		SalaryCurrency: vacancy.SalaryCurrency,
		OrganizationID: vacancy.OrganizationID,
		CategoryID:     vacancy.CategoryID,
		Details:        detailResponses,
		Organization:   *organization,
		Country:        vacancy.Country,
	}, nil
}

func (s *VacancyService) UpdateVacancy(ctx context.Context, req dto.UpdateVacancyRequest) (*dto.UpdateVacancyResponse, error) {
	err := s.vacancy.Update(ctx, &model.Vacancy{
		ID:             req.Vacancy.ID,
		Title:          req.Vacancy.Title,
		Description:    req.Vacancy.Description,
		SalaryFrom:     req.Vacancy.SalaryFrom,
		SalaryTo:       req.Vacancy.SalaryTo,
		SalaryExact:    req.Vacancy.SalaryExact,
		SalaryType:     req.Vacancy.SalaryType,
		SalaryCurrency: req.Vacancy.SalaryCurrency,
		OrganizationID: req.Vacancy.OrganizationID,
		CategoryID:     req.Vacancy.CategoryID,
		Country:        req.Vacancy.Country,
	})
	if err != nil {
		return nil, err
	}

	for _, detail := range req.Vacancy.Details {
		err := s.detail.Update(ctx, &model.VacancyDetail{
			ID:        detail.ID,
			GroupName: detail.GroupName,
			Name:      detail.Name,
			Value:     detail.Value,
			IconURL:   &detail.IconURL,
			VacancyID: req.Vacancy.ID,
		})
		if err != nil {
			return nil, err
		}
	}

	return &dto.UpdateVacancyResponse{Vacancy: req.Vacancy}, nil
}

func (s *VacancyService) DeleteVacancy(ctx context.Context, id int64) error {
	return s.vacancy.Delete(ctx, id)
}

func (s *VacancyService) ListVacancies(ctx context.Context, req dto.ListVacancyRequest) (*dto.ListVacancyResponse, error) {
	vacancies, total, err := s.vacancy.List(ctx, req)
	if err != nil {
		return nil, err
	}

	var responseVacancies []dto.Vacancy
	for _, v := range vacancies {
		details, _ := s.detail.GetByVacancyID(ctx, v.ID)
		var detailResponses []dto.VacancyDetailResponse
		for _, d := range details {
			detailResponses = append(detailResponses, dto.VacancyDetailResponse{
				ID:        d.ID,
				GroupName: d.GroupName,
				Name:      d.Name,
				Value:     d.Value,
				IconURL:   *d.IconURL,
				VacancyID: d.VacancyID,
			})
		}

		organization, err := s.organization.GetOrganization(int(v.OrganizationID))
		if err != nil {
			return nil, err
		}

		responseVacancies = append(responseVacancies, dto.Vacancy{
			ID:             v.ID,
			Title:          v.Title,
			Description:    v.Description,
			SalaryFrom:     v.SalaryFrom,
			SalaryTo:       v.SalaryTo,
			SalaryExact:    v.SalaryExact,
			SalaryType:     v.SalaryType,
			SalaryCurrency: v.SalaryCurrency,
			OrganizationID: v.OrganizationID,
			CategoryID:     v.CategoryID,
			Details:        detailResponses,
			Organization:   *organization,
			Country:        v.Country,
		})
	}

	return &dto.ListVacancyResponse{Vacancie: responseVacancies, Total: total}, nil
}
