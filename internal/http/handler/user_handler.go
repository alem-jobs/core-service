package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/aidosgal/alem.core-service/internal/dto"
	"github.com/aidosgal/alem.core-service/internal/http/middleware"
	"github.com/aidosgal/alem.core-service/internal/lib"
	"github.com/aidosgal/alem.core-service/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := lib.ParseJSON(r, &req); err != nil {
		lib.WriteError(w, http.StatusBadRequest, err)
		return
	}

	res, err := h.userService.Register(req)
	if err != nil {
		lib.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	lib.WriteJSON(w, http.StatusCreated, res)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := lib.ParseJSON(r, &req); err != nil {
		lib.WriteError(w, http.StatusBadRequest, err)
		return
	}

	res, err := h.userService.Login(req)
	if err != nil {
		lib.WriteError(w, http.StatusUnauthorized, err)
		return
	}

	lib.WriteJSON(w, http.StatusOK, res)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	profile, err := h.userService.GetProfile(userID)
	if err != nil {
		lib.WriteError(w, http.StatusNotFound, err)
		return
	}

	lib.WriteJSON(w, http.StatusOK, profile)
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
    userID, ok := middleware.GetUserID(r)
    if !ok {
        lib.WriteError(w, http.StatusUnauthorized, fmt.Errorf("Unauthorized"))
    }

	profile, err := h.userService.GetProfile(int(userID))
	if err != nil {
		lib.WriteError(w, http.StatusNotFound, err)
		return
	}

	lib.WriteJSON(w, http.StatusOK, profile)
}
