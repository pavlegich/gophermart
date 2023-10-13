package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/gophermart/internal/domains/user"
	errs "github.com/pavlegich/gophermart/internal/errors"
	"github.com/pavlegich/gophermart/internal/infra/config"
	"github.com/pavlegich/gophermart/internal/infra/hash"
	"github.com/pavlegich/gophermart/internal/infra/logger"
)

type Handler struct {
	Config  *config.Config
	Service user.Service
}

func NewHandler(cfg *config.Config, s user.Service) *Handler {
	return &Handler{
		Config:  cfg,
		Service: s,
	}
}

// Route регистрирует обработчики и мидлвары в роутере
func (h *Handler) BuildRoute(ctx context.Context) *chi.Mux {
	r := chi.NewRouter()

	r.Post("/register", h.HandleRegister)
	r.Post("/login", h.HandleLogin)

	return r
}

// HandleRegister регистрирует нового пользователя
func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req user.User
	var buf bytes.Buffer

	if _, err := buf.ReadFrom(r.Body); err != nil {
		logger.Log.Info("HandleRegister: read body failed")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(buf.Bytes(), &req); err != nil {
		logger.Log.Info("HandleRegister: unmarshal failed")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.Service.Register(ctx, &req); err != nil {
		if errors.Is(err, errs.ErrLoginBusy) {
			logger.Log.Info("HandleRegister: login is already busy")
			w.WriteHeader(http.StatusConflict)
			return
		} else {
			logger.Log.Info("HandleRegister: register user failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	ctx = context.WithValue(ctx, "Login", req.Login)
	token, err := hash.BuildJWTString(ctx)
	if err != nil {
		logger.Log.Info("HandleRegister: build token failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{
		Name:     "auth",
		Value:    token,
		Secure:   true,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req user.User
	var buf bytes.Buffer

	if _, err := buf.ReadFrom(r.Body); err != nil {
		logger.Log.Info("HandleLogin: read body failed")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(buf.Bytes(), &req); err != nil {
		logger.Log.Info("HandleLogin: unmarshal failed")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.Service.Login(ctx, &req); err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			logger.Log.Info("HandleLogin: user is not found")
			w.WriteHeader(http.StatusUnauthorized)
		} else if errors.Is(err, errs.ErrPasswordNotMatch) {
			logger.Log.Info("HandleLogin: passwords do not match")
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			logger.Log.Info("HandleLogin: login failed")
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	ctx = context.WithValue(ctx, "Login", req.Login)
	token, err := hash.BuildJWTString(ctx)
	if err != nil {
		logger.Log.Info("HandleLogin: build token failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{
		Name:     "auth",
		Value:    token,
		Secure:   true,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	w.WriteHeader(http.StatusOK)
}
