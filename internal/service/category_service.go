package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/aidosgal/alem.core-service/internal/dto"
	"github.com/aidosgal/alem.core-service/internal/model"
	"github.com/aidosgal/alem.core-service/internal/repository"
)

type CategoryService struct {
    log *slog.Logger
	repo *repository.CategoryRepository
}

func NewCategoryService(repo *repository.CategoryRepository, log *slog.Logger) *CategoryService {
    return &CategoryService{repo: repo, log: log}
}

func (s *CategoryService) CreateCategory(ctx context.Context, req dto.CreateCategory) (*dto.CategoryResponse, error) {
	var left, right, depth int
	if req.ParentID != nil {
		parent, err := s.repo.FindByID(ctx, *req.ParentID)
		if err != nil || parent == nil {
			return nil, errors.New("invalid parent ID")
		}
		left = parent.Right
		right = parent.Right + 1
		depth = parent.Depth + 1
	} else {
		left = 1
		right = 2
		depth = 1
	}

	category := &model.Category{
		Name:     req.Name,
		ParentID: req.ParentID,
		Left:     left,
		Right:    right,
		Depth:    depth,
	}

	err := s.repo.Insert(ctx, category)
	if err != nil {
		return nil, err
	}

	return &dto.CategoryResponse{
		ID:       category.ID,
		Name:     category.Name,
		ParentID: category.ParentID,
		Left:     category.Left,
		Right:    category.Right,
		Depth:    category.Depth,
	}, nil
}

func (s *CategoryService) GetCategoryByID(ctx context.Context, id int) (*dto.CategoryResponse, error) {
	category, err := s.repo.FindByID(ctx, id)
	if err != nil || category == nil {
		return nil, errors.New("category not found")
	}
	return &dto.CategoryResponse{
		ID:       category.ID,
		Name:     category.Name,
		ParentID: category.ParentID,
		Left:     category.Left,
		Right:    category.Right,
		Depth:    category.Depth,
	}, nil
}

func (s *CategoryService) UpdateCategory(ctx context.Context, req dto.UpdateCategory) error {
	category, err := s.repo.FindByID(ctx, req.ID)
	if err != nil || category == nil {
		return errors.New("category not found")
	}

	category.Name = req.Name

	return s.repo.Update(ctx, category)
}

func (s *CategoryService) DeleteCategory(ctx context.Context, id int) error {
	category, err := s.repo.FindByID(ctx, id)
	if err != nil || category == nil {
		return errors.New("category not found")
	}

	return s.repo.Delete(ctx, id)
}

func (s *CategoryService) GetCategoryTree(ctx context.Context) ([]dto.CategoryResponse, error) {
	categories, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var result []dto.CategoryResponse
	for _, c := range categories {
		result = append(result, dto.CategoryResponse{
			ID:       c.ID,
			Name:     c.Name,
			ParentID: c.ParentID,
			Left:     c.Left,
			Right:    c.Right,
			Depth:    c.Depth,
		})
	}
	return result, nil
}
