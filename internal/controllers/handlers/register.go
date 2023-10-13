package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/pavlegich/gophermart/internal/domains/user"
	errs "github.com/pavlegich/gophermart/internal/errors"
	"github.com/pavlegich/gophermart/internal/infra/hash"
	"github.com/pavlegich/gophermart/internal/infra/logger"
)

// HandleRegister регистрирует нового пользователя
func (c *Controller) HandleRegister(w http.ResponseWriter, r *http.Request) {
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

	if err := c.User.Register(ctx, &req); err != nil {
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
