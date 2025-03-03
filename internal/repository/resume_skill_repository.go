package repository

import (
	"context"
	"log/slog"
	"database/sql"

	"github.com/aidosgal/alem.core-service/internal/model"
)

type ResumeSkillRepository struct {
	log *slog.Logger
	db *sql.DB
}

func NewResumeSkillRepository(log *slog.Logger, db *sql.DB) *ResumeSkillRepository {
	return &ResumeSkillRepository{
		log: log,
		db: db,
	}
}

func (r *ResumeSkillRepository) CreateResumeSkill(ctx context.Context, resume_skill model.ResumeSkill) (*model.ResumeSkill, error) {
	query := `
        INSERT INTO resume_skills (resume_id, skill)
        VALUES ($1, $2) RETURNING id`

	row := r.db.QueryRowContext(ctx, query, resume_skill.ResumeId, resume_skill.Skill)
	var resume_skill_id int
	err := row.Scan(&resume_skill_id)
	if err != nil {
		return nil, err
	}

	return r.GetResumeSkill(ctx, resume_skill_id)
}

func (r *ResumeSkillRepository) GetResumeSkill(ctx context.Context, resume_skill_id int) (*model.ResumeSkill, error) {
	query := `
        SELECT id, resume_id, skill
        FROM resume_skills
        WHERE id = $1
    `

	resume := &model.ResumeSkill{}
	err := r.db.QueryRowContext(ctx, query, resume_skill_id).Scan(
		&resume.Id, &resume.ResumeId, &resume.Skill,
	)
	if err != nil {
		return nil, err
	}

	return resume, nil
}

func (r *ResumeSkillRepository)  ListResumeSkills(ctx context.Context, resume_id int) ([]*model.ResumeSkill, error) {
	query := `
        SELECT id, resume_id, skill
        FROM resume_skills
        WHERE resume_id = $1
    `

	rows, err := r.db.QueryContext(ctx, query, resume_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resumes []*model.ResumeSkill
	for rows.Next() {
		resume := &model.ResumeSkill{}
		if err := rows.Scan(
			&resume.Id, &resume.ResumeId, &resume.Skill,
		); err != nil {
			return nil, err
		}
		resumes = append(resumes, resume)
	}

	return resumes, nil
}
