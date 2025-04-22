package app

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aidosgal/alem.core-service/internal/config"
	"github.com/aidosgal/alem.core-service/internal/http/handler"
	auth "github.com/aidosgal/alem.core-service/internal/http/middleware"
	"github.com/aidosgal/alem.core-service/internal/repository"
	"github.com/aidosgal/alem.core-service/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
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
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	publicDir := filepath.Join(cwd, "public")
	if err = os.MkdirAll(filepath.Join(publicDir, "files"), 0755); err != nil {
		return err
	}

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.URLFormat)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

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

	vacancyRepository := repository.NewVacancyRepository(s.log, db)
	vacancyDetailRepository := repository.NewVacancyDetailRepository(s.log, db)
	vacancyService := service.NewVacancyService(s.log, vacancyRepository, vacancyDetailRepository)
	vacancyHandler := handler.NewVacancyHandler(s.log, vacancyService)

	resumeRepository := repository.NewResumeRepository(s.log, db)
	resumeExperienceRepository := repository.NewResumeExperienceRepository(s.log, db)
	resumeSkillRepository := repository.NewResumeSkillRepository(s.log, db)
	resumeService := service.NewResumeService(
		s.log, resumeRepository, resumeSkillRepository, resumeExperienceRepository)
	resumeHandler := handler.NewResumeHandler(s.log, resumeService)

	messageRepo := repository.NewMessageRepository(db)
	chatService := service.NewChatService(messageRepo, publicDir)
	wsHandler := handler.NewWebSocketHandler(chatService)

	fileServer(router, "/files", http.Dir(filepath.Join(publicDir, "files")))

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
		apiRouter.Route("/vacancy", func(vacancyRouter chi.Router) {
			vacancyRouter.Use(auth.AuthMiddleware)
			vacancyRouter.Post("/", vacancyHandler.CreateVacancy)
			vacancyRouter.Get("/", vacancyHandler.ListVacancies)
			vacancyRouter.Get("/{id}", vacancyHandler.GetVacancy)
			vacancyRouter.Put("/{id}", vacancyHandler.UpdateVacancy)
		})
		apiRouter.Route("/resumes", func(resumeRouter chi.Router) {
			resumeRouter.Use(auth.AuthMiddleware)
			resumeRouter.Post("/", resumeHandler.CreateResume)
			resumeRouter.Get("/", resumeHandler.ListResume)
			resumeRouter.Get("/{resume_id}", resumeHandler.GetResume)
		})
		apiRouter.Route("/messages", func(wsRouter chi.Router) {
			wsRouter.Use(auth.AuthMiddleware)
			wsRouter.Get("/ws", wsHandler.HandleWebSocket)
			wsRouter.Post("/", wsHandler.SendMessage)
			wsRouter.Get("/", wsHandler.GetMessages)
			wsRouter.Get("/rooms", wsHandler.GetRooms)
		})
	})

	return http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", s.cfg.Port), router)
}

func fileServer(r chi.Router, path string, root http.FileSystem) {
	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
