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

type OrganizationHandler struct {
	log     *slog.Logger
	service *service.OrganizationService
}

func NewOrganizationHandler(log *slog.Logger, service *service.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{
		log:     log,
		service: service,
	}
}

func (h *OrganizationHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var req dto.Organization
	if err := lib.ParseJSON(r, &req); err != nil {
		h.log.Warn("Failed to parse request body", slog.Any("error", err))
		lib.WriteError(w, http.StatusBadRequest, err)
		return
	}

	org, err := h.service.CreateOrganization(&req)
	if err != nil {
		h.log.Error("Failed to create organization", slog.Any("error", err))
		lib.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	h.log.Info("Organization created successfully", slog.Int("id", org.Id))
	lib.WriteJSON(w, http.StatusCreated, org)
}

func (h *OrganizationHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.log.Warn("Invalid organization ID", slog.String("id", idStr))
		lib.WriteError(w, http.StatusBadRequest, err)
		return
	}

	org, err := h.service.GetOrganization(id)
	if err != nil {
		h.log.Error("Failed to retrieve organization", slog.Int("id", id), slog.Any("error", err))
		lib.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	if org == nil {
		h.log.Warn("Organization not found", slog.Int("id", id))
		lib.WriteError(w, http.StatusNotFound, nil)
		return
	}

	h.log.Info("Organization retrieved successfully", slog.Int("id", id))
	lib.WriteJSON(w, http.StatusOK, org)
}

func (h *OrganizationHandler) GetAllOrganizations(w http.ResponseWriter, r *http.Request) {
	orgs, err := h.service.GetAllOrganizations()
	if err != nil {
		h.log.Error("Failed to retrieve organizations", slog.Any("error", err))
		lib.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	h.log.Info("Organizations retrieved successfully", slog.Int("count", len(orgs)))
	lib.WriteJSON(w, http.StatusOK, orgs)
}

