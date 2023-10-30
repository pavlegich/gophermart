package http

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/gophermart/internal/domains/order"
	repo "github.com/pavlegich/gophermart/internal/domains/order/repository"
	errs "github.com/pavlegich/gophermart/internal/errors"
	"github.com/pavlegich/gophermart/internal/infra/config"
	"github.com/pavlegich/gophermart/internal/infra/logger"
	"github.com/pavlegich/gophermart/internal/utils"
	"go.uber.org/zap"
)

type OrderHandler struct {
	Config  *config.Config
	Service order.Service
	Jobs    chan order.Order
}

type responseOrder struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float32 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
}

// Activate активирует обработчик запросов для заказов
func Activate(ctx context.Context, r *chi.Mux, cfg *config.Config, db *sql.DB) {
	s := order.NewOrderService(repo.NewOrderRepo(db))
	newHandler(ctx, r, cfg, s)
}

// newHandler инициализирует обработчик запросов для заказов
func newHandler(ctx context.Context, r *chi.Mux, cfg *config.Config, s order.Service) {
	jobs := make(chan order.Order)
	h := OrderHandler{
		Config:  cfg,
		Service: s,
		Jobs:    jobs,
	}
	r.Post("/api/user/orders", h.HandleOrdersUpload)
	r.Get("/api/user/orders", h.HandleOrdersGet)

	for w := 1; w <= cfg.RateLimit; w++ {
		go workerRequestAccrual(ctx, &h, h.Jobs)
	}
	go workerCheckOrders(ctx, &h)
}

// HandleOrdersGet передаёт список заказов пользователя
func (h *OrderHandler) HandleOrdersGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := utils.GetUserIDFromContext(ctx)
	idString := strconv.Itoa(userID)
	if err != nil {
		logger.Log.With(zap.String("user_id", idString)).Error("HandleOrdersGet: get user id from context failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ordersList, err := h.Service.List(ctx, userID)
	if err != nil {
		if errors.Is(err, errs.ErrOrdersNotFound) {
			logger.Log.With(zap.String("user_id", idString)).Error("HandleOrdersGet: orders not found for this user",
				zap.Error(err))
			w.WriteHeader(http.StatusNoContent)
		} else {
			logger.Log.With(zap.String("user_id", idString)).Error("HandleOrdersGet: get orders list failed",
				zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	resp := make([]responseOrder, 0)
	for _, o := range ordersList {
		tmp := responseOrder{
			Number:     o.Number,
			Status:     o.Status,
			Accrual:    o.Accrual,
			UploadedAt: o.CreatedAt.Format(time.RFC3339),
		}
		resp = append(resp, tmp)
	}

	respJSON, err := json.Marshal(resp)
	if err != nil {
		logger.Log.With(zap.String("user_id", idString)).Error("HandleOrdersGet: response marshal failed",
			zap.Error(err))
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
		logger.Log.Error("HandleOrdersUpload: read request body failed",
			zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	req.Number = buf.String()

	userID, err := utils.GetUserIDFromContext(ctx)
	idString := strconv.Itoa(userID)
	if err != nil {
		logger.Log.With(zap.String("user_id", idString)).Error("HandleOrdersUpload: get user id from context failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	req.UserID = userID

	if err := h.Service.Create(ctx, &req); err != nil {
		if errors.Is(err, errs.ErrOrderAlreadyUpload) {
			w.WriteHeader(http.StatusOK)
		} else if errors.Is(err, errs.ErrOrderUploadByAnother) {
			w.WriteHeader(http.StatusConflict)
		} else if errors.Is(err, errs.ErrIncorrectNumberFormat) {
			w.WriteHeader(http.StatusUnprocessableEntity)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		logger.Log.With(zap.String("user_id", idString)).Error("HandleOrdersUpload: create new order failed",
			zap.Error(err))
		return
	}

	h.Jobs <- order.Order{
		ID:        req.ID,
		Number:    req.Number,
		UserID:    req.UserID,
		Status:    req.Status,
		CreatedAt: req.CreatedAt,
	}

	w.WriteHeader(http.StatusAccepted)
}
