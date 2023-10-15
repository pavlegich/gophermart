package http

import (
	"bytes"
	"database/sql"
	"errors"
	"net/http"
	"strconv"

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
	r.Post("/api/user/orders", h.HandleOrderUpload)
}

// HandleOrderUpload принимает и обрабатывает номер заказа
func (h *OrderHandler) HandleOrderUpload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req order.Order
	var buf bytes.Buffer

	if _, err := buf.ReadFrom(r.Body); err != nil {
		logger.Log.Info("HandleOrderUpload: read body failed")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	orderNumber, err := strconv.Atoi(buf.String())
	if err != nil {
		logger.Log.Info("HandleOrderUpload: order has incorrect number format")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	req.Number = orderNumber

	ctxValue := ctx.Value(utils.ContextIDKey)
	if ctxValue == nil {
		logger.Log.Info("HandleOrderUpload: get context value failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	userID, ok := ctxValue.(int)
	if !ok {
		logger.Log.Info("HandleOrderUpload: convert context value into integer failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	req.UserID = userID

	if err := h.Service.Upload(ctx, &req); err != nil {
		if errors.Is(err, errs.ErrOrderAlreadyUpload) {
			logger.Log.Info("HandleOrderUpload: order already uploaded by this user")
			w.WriteHeader(http.StatusOK)
		} else if errors.Is(err, errs.ErrOrderUploadByAnother) {
			logger.Log.Info("HandleOrderUpload: order uploaded by another user")
			w.WriteHeader(http.StatusConflict)
		} else if errors.Is(err, errs.ErrIncorrectNumberFormat) {
			logger.Log.Info("HandleOrderUpload: order has incorrect number format")
			w.WriteHeader(http.StatusUnprocessableEntity)
		} else {
			logger.Log.Info("HandleOrderUpload: order upload failed")
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
