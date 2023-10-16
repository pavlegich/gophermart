package http

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/gophermart/internal/domains/user"
	repo "github.com/pavlegich/gophermart/internal/domains/user/repository"
	errs "github.com/pavlegich/gophermart/internal/errors"
	"github.com/pavlegich/gophermart/internal/infra/config"
	"github.com/pavlegich/gophermart/internal/infra/hash"
	"github.com/pavlegich/gophermart/internal/infra/logger"
)

// UserHandler содержит интерфейсы и данные обработчика для пользователей
type UserHandler struct {
	Config  *config.Config
	Service user.Service
}

// Activate активирует обработчик запросов для пользователя
func Activate(r *chi.Mux, cfg *config.Config, db *sql.DB) {
	s := user.NewUserService(repo.NewUserRepo(db))
	newHandler(r, cfg, s)
}

// newHandler инициализирует обработчик запросов для пользователя
func newHandler(r *chi.Mux, cfg *config.Config, s user.Service) {
	h := UserHandler{
		Config:  cfg,
		Service: s,
	}
	r.Post("/api/user/register", h.HandleRegister)
	r.Post("/api/user/login", h.HandleLogin)
	r.Post("/api/user/logout", h.HandleLogout)
}

// HandleRegister регистрирует нового пользователя
func (h *UserHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
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

	token, err := hash.BuildJWTString(ctx, req.ID)
	if err != nil {
		logger.Log.Info("HandleRegister: build token failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{
		Name:  "auth",
		Value: token,
		Path:  "/api/user/",
		// Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)
	w.WriteHeader(http.StatusOK)
}

// HandleLogin авторизует пользователя по полученным данным
func (h *UserHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
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

	storedUser, err := h.Service.Login(ctx, &req)
	if err != nil {
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

	token, err := hash.BuildJWTString(ctx, storedUser.ID)
	if err != nil {
		logger.Log.Info("HandleLogin: build token failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{
		Name:  "auth",
		Value: token,
		Path:  "/api/user/",
		// Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)
	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name: "auth",
		Path: "/api/user/",
		// Secure:   true,
		HttpOnly: true,
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusOK)
}
