package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/aidosgal/alem.core-service/internal/model"
)

type CategoryRepository struct {
    log *slog.Logger
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Insert(ctx context.Context, category *model.Category) error {
	query := `
        INSERT INTO categories 
        (name, parent_id, lft, rgt, depth) 
        VALUES ($1, $2, $3, $4, $5) 
        RETURNING id
    `
	return r.db.QueryRowContext(
        ctx, 
        query, 
        category.Name, 
        category.ParentID, 
        category.Left, 
        category.Right, 
        category.Depth,
    ).Scan(&category.ID)
}

func (r *CategoryRepository) FindByID(ctx context.Context, id int) (*model.Category, error) {
	query := `SELECT id, name, parent_id, lft, rgt, depth FROM categories WHERE id = $1`
	var category model.Category
	err := r.db.QueryRowContext(ctx, query, id).Scan(&category.ID, &category.Name, &category.ParentID, &category.Left, &category.Right, &category.Depth)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

func (r *CategoryRepository) Update(ctx context.Context, category *model.Category) error {
	query := `UPDATE categories SET name = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, category.Name, category.ID)
	return err
}

func (r *CategoryRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM categories WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *CategoryRepository) FindAll(ctx context.Context) ([]model.Category, error) {
	query := `SELECT id, name, parent_id, lft, rgt, depth FROM categories ORDER BY lft`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []model.Category
	for rows.Next() {
		var category model.Category
		err := rows.Scan(&category.ID, &category.Name, &category.ParentID, &category.Left, &category.Right, &category.Depth)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}

