package handler

import (
	"net/http"
	"strconv"

	"github.com/aidosgal/alem.core-service/internal/dto"
	"github.com/aidosgal/alem.core-service/internal/lib"
	"github.com/aidosgal/alem.core-service/internal/service"
	"github.com/go-chi/chi/v5"
	"log/slog"
)

type VacancyHandler struct {
	log     *slog.Logger
	service *service.VacancyService
}

func NewVacancyHandler(log *slog.Logger, service *service.VacancyService) *VacancyHandler {
	return &VacancyHandler{
		log:     log,
		service: service,
	}
}

func (h *VacancyHandler) CreateVacancy(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateVacancyRequest
	if err := lib.ParseJSON(r, &req); err != nil {
		h.log.Warn("Failed to parse request body", slog.Any("error", err))
		lib.WriteError(w, http.StatusBadRequest, err)
		return
	}

	vacancy, err := h.service.CreateVacancy(r.Context(), req)
	if err != nil {
		h.log.Error("Failed to create vacancy", slog.Any("error", err))
		lib.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	h.log.Info("Vacancy created successfully", slog.Int64("id", vacancy.Vacancy.ID))
	lib.WriteJSON(w, http.StatusCreated, vacancy)
}

func (h *VacancyHandler) GetVacancy(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.log.Warn("Invalid vacancy ID", slog.String("id", idStr))
		lib.WriteError(w, http.StatusBadRequest, err)
		return
	}

	vacancy, err := h.service.GetVacancyByID(r.Context(), id)
	if err != nil {
		h.log.Error("Failed to retrieve vacancy", slog.Int64("id", id), slog.Any("error", err))
		lib.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	if vacancy == nil {
		h.log.Warn("Vacancy not found", slog.Int64("id", id))
		lib.WriteError(w, http.StatusNotFound, nil)
		return
	}

	h.log.Info("Vacancy retrieved successfully", slog.Int64("id", id))
	lib.WriteJSON(w, http.StatusOK, vacancy)
}

func (h *VacancyHandler) ListVacancies(w http.ResponseWriter, r *http.Request) {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	categoryID, _ := strconv.Atoi(r.URL.Query().Get("category_id"))

	req := dto.ListVacancyRequest{
		Offset:       offset,
		Limit:        limit,
		CategoryID:   int(categoryID),
	}

	vacancies, err := h.service.ListVacancies(r.Context(), req)
	if err != nil {
		h.log.Error("Failed to list vacancies", slog.Any("error", err))
		lib.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	h.log.Info("Vacancies retrieved successfully", slog.Int("total", vacancies.Total))
	lib.WriteJSON(w, http.StatusOK, vacancies)
}

func (h *VacancyHandler) UpdateVacancy(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateVacancyRequest
	if err := lib.ParseJSON(r, &req); err != nil {
		h.log.Warn("Failed to parse request body", slog.Any("error", err))
		lib.WriteError(w, http.StatusBadRequest, err)
		return
	}

	vacancy, err := h.service.UpdateVacancy(r.Context(), req)
	if err != nil {
		h.log.Error("Failed to update vacancy", slog.Any("error", err))
		lib.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	h.log.Info("Vacancy updated successfully", slog.Int64("id", vacancy.Vacancy.ID))
	lib.WriteJSON(w, http.StatusOK, vacancy)
}

func (h *VacancyHandler) DeleteVacancy(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.log.Warn("Invalid vacancy ID", slog.String("id", idStr))
		lib.WriteError(w, http.StatusBadRequest, err)
		return
	}

	err = h.service.DeleteVacancy(r.Context(), id)
	if err != nil {
		h.log.Error("Failed to delete vacancy", slog.Int64("id", id), slog.Any("error", err))
		lib.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	h.log.Info("Vacancy deleted successfully", slog.Int64("id", id))
	lib.WriteJSON(w, http.StatusNoContent, nil)
}
