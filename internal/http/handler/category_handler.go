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

type CategoryHandler struct {
	log     *slog.Logger
	service *service.CategoryService
}

func NewCategoryHandler(log *slog.Logger, service *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		log:     log,
		service: service,
	}
}

// CreateCategory handles creating a new category
func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateCategory
	if err := lib.ParseJSON(r, &req); err != nil {
		h.log.Warn("Failed to parse request body", slog.Any("error", err))
		lib.WriteError(w, http.StatusBadRequest, err)
		return
	}

	category, err := h.service.CreateCategory(r.Context(), req)
	if err != nil {
		h.log.Error("Failed to create category", slog.Any("error", err))
		lib.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	h.log.Info("Category created successfully", slog.Int("id", category.ID))
	lib.WriteJSON(w, http.StatusCreated, category)
}

// GetCategoryByID handles retrieving a category by ID
func (h *CategoryHandler) GetCategoryByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.log.Warn("Invalid category ID", slog.String("id", idStr))
		lib.WriteError(w, http.StatusBadRequest, err)
		return
	}

	category, err := h.service.GetCategoryByID(r.Context(), id)
	if err != nil {
		h.log.Warn("Category not found", slog.Int("id", id))
		lib.WriteError(w, http.StatusNotFound, err)
		return
	}

	h.log.Info("Category retrieved successfully", slog.Int("id", id))
	lib.WriteJSON(w, http.StatusOK, category)
}

// UpdateCategory handles updating a category
func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateCategory
	if err := lib.ParseJSON(r, &req); err != nil {
		h.log.Warn("Failed to parse request body", slog.Any("error", err))
		lib.WriteError(w, http.StatusBadRequest, err)
		return
	}

	err := h.service.UpdateCategory(r.Context(), req)
	if err != nil {
		h.log.Warn("Failed to update category", slog.Any("error", err))
		lib.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	h.log.Info("Category updated successfully", slog.Int("id", req.ID))
	lib.WriteJSON(w, http.StatusOK, map[string]string{"message": "Category updated successfully"})
}

// DeleteCategory handles deleting a category
func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.log.Warn("Invalid category ID", slog.String("id", idStr))
		lib.WriteError(w, http.StatusBadRequest, err)
		return
	}

	err = h.service.DeleteCategory(r.Context(), id)
	if err != nil {
		h.log.Warn("Failed to delete category", slog.Int("id", id))
		lib.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	h.log.Info("Category deleted successfully", slog.Int("id", id))
	lib.WriteJSON(w, http.StatusOK, map[string]string{"message": "Category deleted successfully"})
}

// GetCategoryTree handles retrieving the category tree
func (h *CategoryHandler) GetCategoryTree(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.GetCategoryTree(r.Context())
	if err != nil {
		h.log.Error("Failed to retrieve category tree", slog.Any("error", err))
		lib.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	h.log.Info("Category tree retrieved successfully", slog.Int("count", len(categories)))
	lib.WriteJSON(w, http.StatusOK, categories)
}
