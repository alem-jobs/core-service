package service

import (
	"context"
	"log/slog"

	"github.com/aidosgal/alem.core-service/internal/model"
	"github.com/aidosgal/alem.core-service/internal/repository"
)

type ResumeService struct {
	log        *slog.Logger
	resume     *repository.ResumeRepository
	skill      *repository.ResumeSkillRepository
	experience *repository.ResumeExperienceRepository
	category   *CategoryService
}

func NewResumeService(
	log *slog.Logger,
	resume *repository.ResumeRepository,
	skill *repository.ResumeSkillRepository,
	experience *repository.ResumeExperienceRepository,
	category *CategoryService,
) *ResumeService {
	return &ResumeService{
		log:        log,
		resume:     resume,
		skill:      skill,
		experience: experience,
		category:   category,
	}
}

func (s *ResumeService) CreateResume(ctx context.Context, req *model.Resume) (*model.Resume, error) {
	resume, err := s.resume.CreateResume(ctx, *req)
	if err != nil {
		return nil, err
	}

	for _, skill := range req.Skills {
		_, err := s.skill.CreateResumeSkill(ctx, *skill)
		if err != nil {
			return nil, err
		}
	}

	for _, experience := range req.Experiences {
		_, err := s.experience.CreateResumeExperience(ctx, experience)
		if err != nil {
			return nil, err
		}
	}

	return s.GetResume(ctx, resume.Id)
}

func (s *ResumeService) GetResume(ctx context.Context, id int) (*model.Resume, error) {
	resume, err := s.resume.GetResumeByID(ctx, id)
	if err != nil {
		return nil, err
	}

	resume_skills, err := s.skill.ListResumeSkills(ctx, id)
	if err != nil {
		return nil, err
	}

	resume.Skills = resume_skills

	resume_experiences, err := s.experience.ListResumeExperiencesByResumeID(ctx, id)
	if err != nil {
		return nil, err
	}

	resume.Experiences = resume_experiences

	category, err := s.category.GetCategoryByID(ctx, resume.CategoryId)
	if err != nil {
		return nil, err
	}

	resume.Category = category

	return &resume, nil
}

func (s *ResumeService) ListResumes(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*model.Resume, error) {
	resumes, err := s.resume.ListResumes(ctx, filters, limit, offset)
	if err != nil {
		return nil, err
	}

	var detailed_resumes []*model.Resume
	for _, resume := range resumes {
		detailed_resume, err := s.GetResume(ctx, resume.Id)
		if err != nil {
			return nil, err
		}

		detailed_resumes = append(detailed_resumes, detailed_resume)
	}

	return detailed_resumes, nil
}
