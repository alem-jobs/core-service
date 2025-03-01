package repository

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/aidosgal/alem.core-service/internal/model"
)

type VacancyDetailRepository struct {
	log *slog.Logger
	db  *sql.DB
}

func NewVacancyDetailRepository(log *slog.Logger, db *sql.DB) *VacancyDetailRepository {
	return &VacancyDetailRepository{
		log: log,
		db:  db,
	}
}

func (r *VacancyDetailRepository) Create(ctx context.Context, detail *model.VacancyDetail) (int64, error) {
	query := `INSERT INTO vacancy_details (group_name, name, value, icon_url, vacancy_id) 
			VALUES ($1, $2, $3, $4, $5) RETURNING id`
	var id int64
	err := r.db.QueryRowContext(ctx, query, detail.GroupName, detail.Name, detail.Value, detail.IconURL, detail.VacancyID).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *VacancyDetailRepository) GetByVacancyID(ctx context.Context, vacancyID int64) ([]model.VacancyDetail, error) {
	query := `SELECT id, group_name, name, value, icon_url, vacancy_id FROM vacancy_details WHERE vacancy_id = $1`
	rows, err := r.db.QueryContext(ctx, query, vacancyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var details []model.VacancyDetail
	for rows.Next() {
		var detail model.VacancyDetail
		err := rows.Scan(&detail.ID, &detail.GroupName, &detail.Name, &detail.Value, &detail.IconURL, &detail.VacancyID)
		if err != nil {
			return nil, err
		}
		details = append(details, detail)
	}
	return details, nil
}

func (r *VacancyDetailRepository) Update(ctx context.Context, detail *model.VacancyDetail) error {
	query := `UPDATE vacancy_details SET group_name=$1, name=$2, value=$3, icon_url=$4 WHERE id=$5`
	_, err := r.db.ExecContext(ctx, query, detail.GroupName, detail.Name, detail.Value, detail.IconURL, detail.ID)
	return err
}

func (r *VacancyDetailRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM vacancy_details WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
