package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/aidosgal/alem.core-service/internal/http/middleware"
	"github.com/aidosgal/alem.core-service/internal/lib"
	"github.com/aidosgal/alem.core-service/internal/model"
	"github.com/aidosgal/alem.core-service/internal/service"
	"github.com/go-chi/chi/v5"
)

type ResumeHandler struct {
	log *slog.Logger
	service *service.ResumeService
}

func NewResumeHandler(log *slog.Logger, service *service.ResumeService) *ResumeHandler {
	return &ResumeHandler{
		log: log,
		service: service,
	}
}

func (h *ResumeHandler) CreateResume(w http.ResponseWriter, r *http.Request) {
	resume := &model.Resume{}

	err := lib.ParseJSON(r, &resume)
	if err != nil {
		lib.WriteError(w, http.StatusBadRequest, err)
		return
	}

	user_id, ok := middleware.GetUserID(r)
	if !ok {
		lib.WriteError(w, http.StatusUnauthorized, fmt.Errorf("Unauthorized"))
		return
	}

	resume.UserId = int(user_id)

	response, err := h.service.CreateResume(r.Context(), resume)
	if err != nil {
		lib.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	lib.WriteJSON(w, http.StatusCreated, map[string]interface{}{"resume": response})
	return
}

func (h *ResumeHandler) ListResume(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	user_id_str := query.Get("user_id")
	user_id := 0
	user_id, _ = strconv.Atoi(user_id_str)

	category_id_str := query.Get("category_id")
	category_id := 0
	category_id, _ = strconv.Atoi(category_id_str)
	
	limit := 10
	offset := 0
	
	limit_str := query.Get("limit")
	limit, _ = strconv.Atoi(limit_str)

	offset_str := query.Get("offset")
	offset, _ = strconv.Atoi(offset_str)
	
	filters := map[string]interface{}{
		"user_id": user_id,
		"category_id": category_id,
	}

	resumes, err := h.service.ListResumes(r.Context(), filters, limit, offset)
	if err != nil {
		lib.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	lib.WriteJSON(w, http.StatusOK, map[string]interface{}{"resumes": resumes})
	return
}

func (h *ResumeHandler) GetResume(w http.ResponseWriter, r *http.Request) {
	resume_id_str := chi.URLParam(r, "resume_id")
	resume_id, err := strconv.Atoi(resume_id_str)
	if err != nil {
		lib.WriteError(w, http.StatusBadRequest, err)
		return
	}

	resume, err := h.service.GetResume(r.Context(), resume_id)
	if err != nil {
		lib.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	lib.WriteJSON(w, http.StatusOK, map[string]interface{}{"resume": resume})
	return
}
