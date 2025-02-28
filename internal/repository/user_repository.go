package repository

import (
	"database/sql"
	"log/slog"

	"github.com/aidosgal/alem.core-service/internal/model"
)

type UserRepository struct {
	log *slog.Logger
	db  *sql.DB
}

func NewUserRepository(log *slog.Logger, db *sql.DB) *UserRepository {
	return &UserRepository{
		log: log,
		db:  db,
	}
}

func (r *UserRepository) CreateUser(user *model.User) (int64, error) {
	query := `INSERT INTO users (name, organization_id, phone, password, avatar_url, balance, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())`
	
	result, err := r.db.Exec(query, user.Name, user.OrganizationId, user.Phone, user.Password, user.AvatarURL, user.Balance)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}


func (r *UserRepository) GetUserByPhone(phone string) (*model.User, error) {
	query := `SELECT id, name, organization_id, phone, password, avatar_url, balance, created_at, updated_at FROM users WHERE phone = ?`
	row := r.db.QueryRow(query, phone)

	var user model.User
	err := row.Scan(&user.Id, &user.Name, &user.OrganizationId, &user.Phone, &user.Password, &user.AvatarURL, &user.Balance, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByID(id int) (*model.User, error) {
	query := `SELECT id, name, organization_id, phone, password, avatar_url, balance, created_at, updated_at FROM users WHERE id = ?`
	row := r.db.QueryRow(query, id)

	var user model.User
	err := row.Scan(&user.Id, &user.Name, &user.OrganizationId, &user.Phone, &user.Password, &user.AvatarURL, &user.Balance, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateUser(user *model.User) error {
	query := `UPDATE users SET name = ?, phone = ?, avatar_url = ?, balance = ?, updated_at = NOW() WHERE id = ?`
	_, err := r.db.Exec(query, user.Name, user.Phone, user.AvatarURL, user.Balance, user.Id)
	return err
}

func (r *UserRepository) ListUsers() ([]model.User, error) {
	query := `SELECT id, name, organization_id, phone, avatar_url, balance, created_at, updated_at FROM users`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		err := rows.Scan(&user.Id, &user.Name, &user.OrganizationId, &user.Phone, &user.AvatarURL, &user.Balance, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

