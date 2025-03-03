package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/aidosgal/alem.core-service/internal/model"
)

type ResumeExperienceRepository struct {
	log *slog.Logger
	db  *sql.DB
}

func NewResumeExperienceRepository(
	log *slog.Logger,
	db *sql.DB,
) *ResumeExperienceRepository {
	return &ResumeExperienceRepository{
		log: log,
		db:  db,
	}
}

func (r *ResumeExperienceRepository) CreateResumeExperience(ctx context.Context, resumeExperience *model.ResumeExperience) (*model.ResumeExperience, error) {
	query := `
		INSERT INTO resume_experiences (
			resume_id, oraganization_name, category_id, description, 
			start_month, start_year, end_month, end_year
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	r.log.Info("creating resume experience", "resume_id", resumeExperience.ResumeId)

	err := r.db.QueryRowContext(
		ctx,
		query,
		resumeExperience.ResumeId,
		resumeExperience.OrganizationName,
		resumeExperience.CategoryId,
		resumeExperience.Description,
		resumeExperience.StartMonth,
		resumeExperience.StartYear,
		resumeExperience.EndMonth,
		resumeExperience.EndYear,
	).Scan(&resumeExperience.Id)

	if err != nil {
		r.log.Error("error creating resume experience", "error", err)
		return nil, fmt.Errorf("failed to create resume experience: %w", err)
	}

	return resumeExperience, nil
}

func (r *ResumeExperienceRepository) GetResumeExperience(ctx context.Context, id int) (*model.ResumeExperience, error) {
	query := `
		SELECT id, resume_id, oraganization_name, category_id, description, 
		       start_month, start_year, end_month, end_year
		FROM resume_experiences
		WHERE id = $1
	`

	r.log.Info("getting resume experience", "id", id)

	var resumeExperience model.ResumeExperience
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&resumeExperience.Id,
		&resumeExperience.ResumeId,
		&resumeExperience.OrganizationName,
		&resumeExperience.CategoryId,
		&resumeExperience.Description,
		&resumeExperience.StartMonth,
		&resumeExperience.StartYear,
		&resumeExperience.EndMonth,
		&resumeExperience.EndYear,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log.Info("resume experience not found", "id", id)
			return nil, fmt.Errorf("resume experience with id %d not found", id)
		}
		r.log.Error("error getting resume experience", "error", err)
		return nil, fmt.Errorf("failed to get resume experience: %w", err)
	}

	return &resumeExperience, nil
}

func (r *ResumeExperienceRepository) UpdateResumeExperience(ctx context.Context, resumeExperience *model.ResumeExperience) (*model.ResumeExperience, error) {
	query := `
		UPDATE resume_experiences
		SET resume_id = $1, 
		    oraganization_name = $2, 
		    category_id = $3, 
		    description = $4, 
		    start_month = $5, 
		    start_year = $6, 
		    end_month = $7, 
		    end_year = $8
		WHERE id = $9
		RETURNING id
	`

	r.log.Info("updating resume experience", "id", resumeExperience.Id)

	result, err := r.db.ExecContext(
		ctx,
		query,
		resumeExperience.ResumeId,
		resumeExperience.OrganizationName,
		resumeExperience.CategoryId,
		resumeExperience.Description,
		resumeExperience.StartMonth,
		resumeExperience.StartYear,
		resumeExperience.EndMonth,
		resumeExperience.EndYear,
		resumeExperience.Id,
	)

	if err != nil {
		r.log.Error("error updating resume experience", "error", err)
		return nil, fmt.Errorf("failed to update resume experience: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.Error("error getting rows affected", "error", err)
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		r.log.Info("resume experience not found", "id", resumeExperience.Id)
		return nil, fmt.Errorf("resume experience with id %d not found", resumeExperience.Id)
	}

	return resumeExperience, nil
}

func (r *ResumeExperienceRepository) DeleteResumeExperience(ctx context.Context, id int) error {
	query := `DELETE FROM resume_experiences WHERE id = $1`

	r.log.Info("deleting resume experience", "id", id)

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.log.Error("error deleting resume experience", "error", err)
		return fmt.Errorf("failed to delete resume experience: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.Error("error getting rows affected", "error", err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		r.log.Info("resume experience not found", "id", id)
		return fmt.Errorf("resume experience with id %d not found", id)
	}

	return nil
}

func (r *ResumeExperienceRepository) ListResumeExperiencesByResumeID(ctx context.Context, resumeID int) ([]*model.ResumeExperience, error) {
	query := `
		SELECT id, resume_id, oraganization_name, category_id, description, 
		       start_month, start_year, end_month, end_year
		FROM resume_experiences
		WHERE resume_id = $1
		ORDER BY start_year DESC, start_month DESC
	`

	r.log.Info("listing resume experiences by resume ID", "resume_id", resumeID)

	rows, err := r.db.QueryContext(ctx, query, resumeID)
	if err != nil {
		r.log.Error("error querying resume experiences", "error", err)
		return nil, fmt.Errorf("failed to query resume experiences: %w", err)
	}
	defer rows.Close()

	var experiences []*model.ResumeExperience
	for rows.Next() {
		var exp model.ResumeExperience
		if err := rows.Scan(
			&exp.Id,
			&exp.ResumeId,
			&exp.OrganizationName,
			&exp.CategoryId,
			&exp.Description,
			&exp.StartMonth,
			&exp.StartYear,
			&exp.EndMonth,
			&exp.EndYear,
		); err != nil {
			r.log.Error("error scanning resume experience row", "error", err)
			return nil, fmt.Errorf("failed to scan resume experience: %w", err)
		}
		experiences = append(experiences, &exp)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("error iterating resume experience rows", "error", err)
		return nil, fmt.Errorf("error iterating resume experiences: %w", err)
	}

	return experiences, nil
}
