package app

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/aidosgal/alem.core-service/internal/config"
	"github.com/aidosgal/alem.core-service/internal/http/handler"
	auth "github.com/aidosgal/alem.core-service/internal/http/middleware"
	"github.com/aidosgal/alem.core-service/internal/repository"
	"github.com/aidosgal/alem.core-service/internal/service"
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

    userRepository := repository.NewUserRepository(s.log, db)
    userService := service.NewUserService(s.log, userRepository)
    userHandler := handler.NewUserHandler(userService)

    organizationRepository := repository.NewOrganizationRepository(s.log, db)
    organizationService := service.NewOrganizationService(s.log, organizationRepository)
    organizationHandler := handler.NewOrganizationHandler(s.log, organizationService)

    categoryRepository := repository.NewCategoryRepository(db)
    categoryService := service.NewCategoryService(categoryRepository, s.log)
    categoryHandler := handler.NewCategoryHandler(s.log, categoryService)

	router.Route("/api/v1", func(apiRouter chi.Router) {
        apiRouter.Route("/auth", func(authRouter chi.Router) {
            authRouter.Post("/login", userHandler.Login)
            authRouter.Post("/register", userHandler.Register)
        })
        apiRouter.Route("/user", func(userRouter chi.Router) {
            userRouter.Use(auth.AuthMiddleware)
            userRouter.Get("/", userHandler.GetProfile)
        })
        apiRouter.Route("/organization", func(organizationRouter chi.Router) {
            organizationRouter.Use(auth.AuthMiddleware)
            organizationRouter.Get("/", organizationHandler.GetAllOrganizations)
            organizationRouter.Get("/{id}", organizationHandler.GetOrganization)
            organizationRouter.Post("/", organizationHandler.CreateOrganization)
        })
        apiRouter.Route("/category", func(categoryRouter chi.Router) {
            categoryRouter.Use(auth.AuthMiddleware)
            categoryRouter.Get("/", categoryHandler.GetCategoryTree)
            categoryRouter.Post("/", categoryHandler.CreateCategory)
            categoryRouter.Get("/{id}", categoryHandler.GetCategoryByID)
        })
    })

	return http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", s.cfg.Port), router)
}
