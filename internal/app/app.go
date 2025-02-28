package app

import (
	"log/slog"
	"fmt"
	"net/http"
	"database/sql"

	"github.com/aidosgal/alem.core-service/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
)

type Server struct {
	cfg *config.Config
	log *slog.Logger
}

func NewServer(cfg *config.Config, log *slog.Logger) *Server {
	return &Server{
		cfg: cfg,
		log: log,
	}
}

func (s *Server) Run() error {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.URLFormat)

	postgresURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		s.cfg.Database.User,
		s.cfg.Database.Password,
		s.cfg.Database.Host,
		s.cfg.Database.Port,
		s.cfg.Database.Name,
		s.cfg.Database.SSLMode,
	)
	
	db, err := sql.Open("postgres", postgresURL)
	if err != nil {
		panic(err)
	}

	_ = db

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	return http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", s.cfg.Port), router)
}
