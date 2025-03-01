package repository

import (
	"database/sql"
	"errors"
	"log/slog"

	"github.com/aidosgal/alem.core-service/internal/model"
)

type OrganizationRepository struct {
	log *slog.Logger
	db  *sql.DB
}

func NewOrganizationRepository(log *slog.Logger, db *sql.DB) *OrganizationRepository {
	return &OrganizationRepository{
		log: log,
		db:  db,
	}
}

func (r *OrganizationRepository) CreateOrganization(org *model.Organization) error {
	query := "INSERT INTO organizations (name, description) VALUES ($1, $2) RETURNING id"
	err := r.db.QueryRow(query, org.Name, org.Description).Scan(&org.Id)
	if err != nil {
		r.log.Error("Failed to create organization", slog.Any("error", err))
		return err
	}

	r.log.Info("Organization created", slog.Int("id", org.Id))
	return nil
}

func (r *OrganizationRepository) GetOrganization(id int) (*model.Organization, error) {
	query := "SELECT id, name, description FROM organizations WHERE id = $1"
	org := &model.Organization{}
	err := r.db.QueryRow(query, id).Scan(&org.Id, &org.Name, &org.Description)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log.Warn("Organization not found", slog.Int("id", id))
			return nil, nil
		}
		r.log.Error("Failed to get organization", slog.Any("error", err))
		return nil, err
	}

	r.log.Info("Organization retrieved", slog.Int("id", org.Id))
	return org, nil
}

func (r *OrganizationRepository) GetAllOrganizations() ([]*model.Organization, error) {
	query := "SELECT id, name, description FROM organizations"
	rows, err := r.db.Query(query)
	if err != nil {
		r.log.Error("Failed to retrieve organizations", slog.Any("error", err))
		return nil, err
	}
	defer rows.Close()

	var organizations []*model.Organization
	for rows.Next() {
		org := &model.Organization{}
		if err := rows.Scan(&org.Id, &org.Name, &org.Description); err != nil {
			r.log.Error("Failed to scan organization row", slog.Any("error", err))
			continue
		}
		organizations = append(organizations, org)
	}

	r.log.Info("Organizations retrieved", slog.Int("count", len(organizations)))
	return organizations, nil
}

