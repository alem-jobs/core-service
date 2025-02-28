package repository

import (
	"database/sql"
	"log/slog"
)

type UserRepository struct {
    log *slog.Logger
    db *sql.DB
}

func NewUserRepository(log *slog.Logger, db *sql.DB) *UserRepository {
    return &UserRepository{
        log: log,
        db: db,
    }
}
