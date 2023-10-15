package http

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/gophermart/internal/domains/order"
	repo "github.com/pavlegich/gophermart/internal/domains/order/repository"
	errs "github.com/pavlegich/gophermart/internal/errors"
	"github.com/pavlegich/gophermart/internal/infra/config"
	"github.com/pavlegich/gophermart/internal/infra/logger"
	"github.com/pavlegich/gophermart/internal/utils"
)

type OrderHandler struct {
	Config  *config.Config
	Service order.Service
}

// Activate активирует обработчик запросов для заказов
func Activate(r *chi.Mux, cfg *config.Config, db *sql.DB) {
	s := order.NewOrderService(repo.NewOrderRepo(db))
	newHandler(r, cfg, s)
}

// newHandler инициализирует обработчик запросов для заказов
func newHandler(r *chi.Mux, cfg *config.Config, s order.Service) {
	h := OrderHandler{
		Config:  cfg,
		Service: s,
	}
	r.Post("/api/user/orders", h.HandleOrdersUpload)
	r.Get("/api/user/orders", h.HandleOrdersGet)
}

// HandleOrdersGet передаёт список заказов пользователя
func (h *OrderHandler) HandleOrdersGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctxValue := ctx.Value(utils.ContextIDKey)
	if ctxValue == nil {
		logger.Log.Info("HandleOrdersGet: get context value failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	userID, ok := ctxValue.(int)
	if !ok {
		logger.Log.Info("HandleOrdersGet: convert context value into integer failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	orders, err := h.Service.List(ctx, userID)
	if err != nil {
		if errors.Is(err, errs.ErrOrdersNotFound) {
			logger.Log.Info("HandleOrdersGet: orders not found for this user")
			w.WriteHeader(http.StatusNoContent)
		} else {
			logger.Log.Info("HandleOrdersUpload: order upload failed")
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	respJSON, err := json.Marshal(orders)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(respJSON))
}

// HandleOrdersUpload принимает и обрабатывает номер заказа
func (h *OrderHandler) HandleOrdersUpload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req order.Order
	var buf bytes.Buffer

	if _, err := buf.ReadFrom(r.Body); err != nil {
		logger.Log.Info("HandleOrdersUpload: read body failed")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req.Number = buf.String()

	ctxValue := ctx.Value(utils.ContextIDKey)
	if ctxValue == nil {
		logger.Log.Info("HandleOrdersUpload: get context value failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	userID, ok := ctxValue.(int)
	if !ok {
		logger.Log.Info("HandleOrdersUpload: convert context value into integer failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	req.UserID = userID

	if err := h.Service.Upload(ctx, &req); err != nil {
		if errors.Is(err, errs.ErrOrderAlreadyUpload) {
			logger.Log.Info("HandleOrdersUpload: order already uploaded by this user")
			w.WriteHeader(http.StatusOK)
		} else if errors.Is(err, errs.ErrOrderUploadByAnother) {
			logger.Log.Info("HandleOrdersUpload: order uploaded by another user")
			w.WriteHeader(http.StatusConflict)
		} else if errors.Is(err, errs.ErrIncorrectNumberFormat) {
			logger.Log.Info("HandleOrdersUpload: order has incorrect number format")
			w.WriteHeader(http.StatusUnprocessableEntity)
		} else {
			logger.Log.Info("HandleOrdersUpload: order upload failed")
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
