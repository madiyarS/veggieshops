package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/models"
	"github.com/veggieshop/backend/internal/repositories"
	"github.com/veggieshop/backend/internal/utils"
)

type CategoryService struct {
	categoryRepo repositories.CategoryRepository
}

func NewCategoryService(cr repositories.CategoryRepository) *CategoryService {
	return &CategoryService{categoryRepo: cr}
}

func (s *CategoryService) GetAll(ctx context.Context, activeOnly bool) ([]*models.Category, error) {
	return s.categoryRepo.GetAll(ctx, activeOnly)
}

func (s *CategoryService) GetByID(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, utils.ErrNotFound
	}
	return category, nil
}

// CategoryPatch частичное обновление категории в админке.
type CategoryPatch struct {
	Name        *string
	Description *string
	IconURL     *string
	Order       *int
	IsActive    *bool
}

func (s *CategoryService) Create(ctx context.Context, c *models.Category) (*models.Category, error) {
	if err := s.categoryRepo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CategoryService) Update(ctx context.Context, id uuid.UUID, patch *CategoryPatch) error {
	cat, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return utils.ErrNotFound
	}
	if patch == nil {
		return nil
	}
	if patch.Name != nil {
		cat.Name = *patch.Name
	}
	if patch.Description != nil {
		cat.Description = *patch.Description
	}
	if patch.IconURL != nil {
		cat.IconURL = *patch.IconURL
	}
	if patch.Order != nil {
		cat.Order = *patch.Order
	}
	if patch.IsActive != nil {
		cat.IsActive = *patch.IsActive
	}
	return s.categoryRepo.Update(ctx, cat)
}

func (s *CategoryService) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.categoryRepo.GetByID(ctx, id); err != nil {
		return utils.ErrNotFound
	}
	return s.categoryRepo.Delete(ctx, id)
}
