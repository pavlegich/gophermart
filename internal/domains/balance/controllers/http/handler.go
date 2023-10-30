package http

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/gophermart/internal/domains/balance"
	repo "github.com/pavlegich/gophermart/internal/domains/balance/repository"
	errs "github.com/pavlegich/gophermart/internal/errors"
	"github.com/pavlegich/gophermart/internal/infra/config"
	"github.com/pavlegich/gophermart/internal/infra/logger"
	"github.com/pavlegich/gophermart/internal/utils"
	"go.uber.org/zap"
)

type (
	BalanceHandler struct {
		Config  *config.Config
		Service balance.Service
	}

	responseBalance struct {
		Current   float32 `json:"current"`
		Withdrawn float32 `json:"withdrawn"`
	}

	requestWithdraw struct {
		Order string  `json:"order"`
		Sum   float32 `json:"sum"`
	}

	responseWithdrawal struct {
		Order       string  `json:"order"`
		Sum         float32 `json:"sum"`
		ProcessedAt string  `json:"processed_at"`
	}
)

// Activate активирует обработчик запросов для балансов
func Activate(r *chi.Mux, cfg *config.Config, db *sql.DB) {
	s := balance.NewBalanceService(repo.NewBalanceRepo(db))
	newHandler(r, cfg, s)
}

// newHandler инициализирует обработчик запросов для балансов
func newHandler(r *chi.Mux, cfg *config.Config, s balance.Service) {
	h := BalanceHandler{
		Config:  cfg,
		Service: s,
	}
	r.Get("/api/user/balance", h.HandleBalanceGet)
	r.Post("/api/user/balance/withdraw", h.HandleBalanceWithdraw)
	r.Get("/api/user/withdrawals", h.HandleWithdrawalsGet)
}

// HandleBalanceGet обрабатывает запрос получения данных о начислениях и списаниях пользователя
func (h *BalanceHandler) HandleBalanceGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := utils.GetUserIDFromContext(ctx)
	idString := strconv.Itoa(userID)
	if err != nil {
		logger.Log.With(zap.String("user_id", idString)).Error("HandleBalanceGet: get user id from context failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	balanceList, err := h.Service.List(ctx, userID)
	if err != nil {
		logger.Log.With(zap.String("user_id", idString)).Error("HandleBalanceGet: balance get failed",
			zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := responseBalance{
		Current:   0,
		Withdrawn: 0,
	}

	for _, b := range balanceList {
		switch b.Action {
		case "ACCRUAL":
			resp.Current += b.Amount
		case "WITHDRAWAL":
			resp.Current -= b.Amount
			resp.Withdrawn += b.Amount
		default:
			logger.Log.With(zap.String("user_id", idString)).Error("HandleBalanceGet: action get failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	respJSON, err := json.Marshal(resp)
	if err != nil {
		logger.Log.With(zap.String("user_id", idString)).Error("HandleBalanceGet: response marshal failed",
			zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(respJSON))
}

// HandleBalanceWithdraw обрабатывает запрос о списании баллов
func (h *BalanceHandler) HandleBalanceWithdraw(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req requestWithdraw
	var buf bytes.Buffer

	userID, err := utils.GetUserIDFromContext(ctx)
	idString := strconv.Itoa(userID)
	if err != nil {
		logger.Log.With(zap.String("user_id", idString)).Error("HandleBalanceWithdraw: get user id from context failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := buf.ReadFrom(r.Body); err != nil {
		logger.Log.With(zap.String("user_id", idString)).Error("HandleBalanceWithdraw: read request body failed",
			zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(buf.Bytes(), &req); err != nil {
		logger.Log.With(zap.String("user_id", idString)).Error("HandleBalanceWithdraw: request unmarshal failed",
			zap.String("body", buf.String()),
			zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	b := balance.Balance{
		Action: "WITHDRAWAL",
		Amount: req.Sum,
		UserID: userID,
		Order:  req.Order,
	}

	if err := h.Service.Withdraw(ctx, &b); err != nil {
		if errors.Is(err, errs.ErrInsufficientFunds) {
			w.WriteHeader(http.StatusPaymentRequired)
		} else if errors.Is(err, errs.ErrIncorrectNumberFormat) {
			w.WriteHeader(http.StatusUnprocessableEntity)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		logger.Log.With(zap.String("user_id", idString)).Error("HandleBalanceWithdraw: withdrawal failed",
			zap.Error(err))
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleWithdrawalsGet обрабатывает запрос получения данных о всех списаниях пользователя
func (h *BalanceHandler) HandleWithdrawalsGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := utils.GetUserIDFromContext(ctx)
	idString := strconv.Itoa(userID)
	if err != nil {
		logger.Log.With(zap.String("user_id", idString)).Error("HandleWithdrawalsGet: get user id from context failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	balanceList, err := h.Service.List(ctx, userID)
	if err != nil {
		if errors.Is(err, errs.ErrOperationsNotFound) {
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		logger.Log.With(zap.String("user_id", idString)).Error("HandleWithdrawalsGet: balance get failed",
			zap.Error(err))
		return
	}

	resp := make([]responseWithdrawal, 0)

	for _, b := range balanceList {
		if b.Action == "WITHDRAWAL" {
			resp = append(resp, responseWithdrawal{
				Order:       b.Order,
				Sum:         b.Amount,
				ProcessedAt: b.CreatedAt.Format(time.RFC3339),
			})
		}
	}

	if len(resp) == 0 {
		logger.Log.With(zap.String("user_id", idString)).Error("HandleWithdrawalsGet: get withdrawals failed",
			zap.Error(errs.ErrWithdrawalsNotFound))
		w.WriteHeader(http.StatusNoContent)
	}

	respJSON, err := json.Marshal(resp)
	if err != nil {
		logger.Log.With(zap.String("user_id", idString)).Error("HandleWithdrawalsGet: response marshal failed",
			zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(respJSON))
}
