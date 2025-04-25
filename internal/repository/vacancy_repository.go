package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/aidosgal/alem.core-service/internal/dto"
	"github.com/aidosgal/alem.core-service/internal/model"
)

type VacancyRepository struct {
	log *slog.Logger
	db  *sql.DB
}

func NewVacancyRepository(log *slog.Logger, db *sql.DB) *VacancyRepository {
	return &VacancyRepository{
		log: log,
		db:  db,
	}
}

func (r *VacancyRepository) Create(ctx context.Context, vacancy *model.Vacancy) (int64, error) {
	query := `INSERT INTO vacancies (title, description, salary_from, salary_to, salary_exact, salary_type, salary_currency, organization_id, category_id, country) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`
	var id int64
	err := r.db.QueryRowContext(ctx, query, vacancy.Title, vacancy.Description, vacancy.SalaryFrom, vacancy.SalaryTo, vacancy.SalaryExact, vacancy.SalaryType, vacancy.SalaryCurrency, vacancy.OrganizationID, vacancy.CategoryID, vacancy.Country).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *VacancyRepository) GetByID(ctx context.Context, id int64) (*model.Vacancy, error) {
	query := `SELECT id, title, description, salary_from, salary_to, salary_exact, salary_type, salary_currency, organization_id, category_id, country FROM vacancies WHERE id = $1`
	var vacancy model.Vacancy
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&vacancy.ID, &vacancy.Title, &vacancy.Description, &vacancy.SalaryFrom, &vacancy.SalaryTo, &vacancy.SalaryExact, &vacancy.SalaryType, &vacancy.SalaryCurrency, &vacancy.OrganizationID, &vacancy.CategoryID, &vacancy.Country,
	)
	if err != nil {
		return nil, err
	}
	return &vacancy, nil
}

func (r *VacancyRepository) Update(ctx context.Context, vacancy *model.Vacancy) error {
	query := `UPDATE vacancies SET title=$1, description=$2, salary_from=$3, salary_to=$4, salary_exact=$5, salary_type=$6, salary_currency=$7, organization_id=$8, category_id=$9, country=$10 WHERE id=$10`
	_, err := r.db.ExecContext(ctx, query, vacancy.Title, vacancy.Description, vacancy.SalaryFrom, vacancy.SalaryTo, vacancy.SalaryExact, vacancy.SalaryType, vacancy.SalaryCurrency, vacancy.OrganizationID, vacancy.CategoryID, vacancy.Country, vacancy.ID)
	return err
}

func (r *VacancyRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM vacancies WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *VacancyRepository) List(ctx context.Context, req dto.ListVacancyRequest) ([]model.Vacancy, int, error) {
	query := `SELECT id, title, description, salary_from, salary_to, salary_exact, salary_type, salary_currency, organization_id, category_id, country FROM vacancies`
	filters := []interface{}{}
	conditions := []string{}

	argIndex := 1

	if req.CategoryID > 0 {
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", argIndex))
		filters = append(filters, req.CategoryID)
		argIndex++
	}
	if req.SalaryFrom > 0 {
		conditions = append(conditions, fmt.Sprintf("salary_from >= $%d", argIndex))
		filters = append(filters, req.SalaryFrom)
		argIndex++
	}
	if req.SalaryTo > 0 {
		conditions = append(conditions, fmt.Sprintf("salary_to <= $%d", argIndex))
		filters = append(filters, req.SalaryTo)
		argIndex++
	}
	if req.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex+1))
		filters = append(filters, "%"+req.Search+"%", "%"+req.Search+"%")
		argIndex += 2
	}

	if len(conditions) > 0 {
		query += " WHERE " + joinConditions(conditions, " AND ")
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	filters = append(filters, req.Limit, req.Offset)

	r.log.Debug("Executing query", slog.String("query", query))

	rows, err := r.db.QueryContext(ctx, query, filters...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var vacancies []model.Vacancy
	for rows.Next() {
		var v model.Vacancy
		if err := rows.Scan(&v.ID, &v.Title, &v.Description, &v.SalaryFrom, &v.SalaryTo, &v.SalaryExact, &v.SalaryType, &v.SalaryCurrency, &v.OrganizationID, &v.CategoryID, &v.Country); err != nil {
			return nil, 0, err
		}
		vacancies = append(vacancies, v)
	}

	var total int

	countQuery := "SELECT COUNT(*) FROM vacancies"
	if len(conditions) > 0 {
		countQuery += " WHERE " + joinConditions(conditions, " AND ")
	}

	if err := r.db.QueryRowContext(ctx, countQuery, filters[:len(filters)-2]...).Scan(&total); err != nil {
		return nil, 0, err
	}

	return vacancies, total, nil
}

func joinConditions(conditions []string, sep string) string {
	result := ""
	for i, cond := range conditions {
		if i > 0 {
			result += " " + sep + " "
		}
		result += cond
	}
	return result
}
