package repository

import (
	"context"
	"database/sql"
	"log/slog"
	"fmt"

	"github.com/aidosgal/alem.core-service/internal/model"
)

type ResumeRepository struct {
	log *slog.Logger
	db  *sql.DB
}

func NewResumeRepository(log *slog.Logger, db *sql.DB) *ResumeRepository {
	return &ResumeRepository{
		log: log,
		db:  db,
	}
}

func (r *ResumeRepository) CreateResume(
    ctx context.Context, 
    resume model.Resume,
) (model.Resume, error) {
	query := `
        INSERT INTO resumes (
            user_id, 
            category_id, 
            description, 
            salary_from, 
            salary_to, 
            salary_period, 
            created_at
        )
        VALUES ($1, $2, $3, $4, $5, $6, NOW())
        RETURNING id, created_at
    `

	row := r.db.QueryRowContext(
        ctx, query, 
        resume.UserId, 
        resume.CategoryId, 
        resume.Description, 
        resume.SalaryFrom, 
        resume.SalaryTo, 
        resume.SalaryPeriod,
    )
	err := row.Scan(&resume.Id, &resume.CreatedAt)
	if err != nil {
		return model.Resume{}, err
	}

	return resume, nil
}

func (r *ResumeRepository) GetResumeByID(ctx context.Context, id int) (model.Resume, error) {
	query := `
        SELECT 
            id, 
            user_id, 
            category_id, 
            description, 
            salary_from, 
            salary_to, 
            salary_period, 
            created_at 
        FROM resumes 
        WHERE id = $1
    `

	var resume model.Resume
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&resume.Id, &resume.UserId, &resume.CategoryId, &resume.Description,
		&resume.SalaryFrom, &resume.SalaryTo, &resume.SalaryPeriod, &resume.CreatedAt,
	)
	if err != nil {
		return model.Resume{}, err
	}

	return resume, nil
}

func (r *ResumeRepository) UpdateResume(ctx context.Context, resume model.Resume) error {
	query := `
        UPDATE resumes SET category_id = $1, description = $2, salary_from = $3, salary_to = $4, salary_period = $5 WHERE id = $6`

	_, err := r.db.ExecContext(ctx, query, resume.CategoryId, resume.Description, resume.SalaryFrom, resume.SalaryTo, resume.SalaryPeriod, resume.Id)
	return err
}

func (r *ResumeRepository) DeleteResume(ctx context.Context, id int) error {
	query := `DELETE FROM resumes WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *ResumeRepository) ListResumes(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]model.Resume, error) {
	query := `SELECT id, user_id, category_id, description, salary_from, salary_to, salary_period, created_at FROM resumes WHERE 1=1`
	args := []interface{}{}
	idx := 1

	if userId, ok := filters["user_id"]; ok {
		query += ` AND user_id = $` + fmt.Sprint(idx)
		args = append(args, userId)
		idx++
	}

	if categoryId, ok := filters["category_id"]; ok {
		query += ` AND category_id = $` + fmt.Sprint(idx)
		args = append(args, categoryId)
		idx++
	}

	query += ` ORDER BY created_at DESC LIMIT $` + fmt.Sprint(idx) + ` OFFSET $` + fmt.Sprint(idx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resumes []model.Resume
	for rows.Next() {
		var resume model.Resume
		if err := rows.Scan(
			&resume.Id, &resume.UserId, &resume.CategoryId, &resume.Description,
			&resume.SalaryFrom, &resume.SalaryTo, &resume.SalaryPeriod, &resume.CreatedAt,
		); err != nil {
			return nil, err
		}
		resumes = append(resumes, resume)
	}

	return resumes, nil
}
