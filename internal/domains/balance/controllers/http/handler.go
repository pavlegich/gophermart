package http

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/gophermart/internal/domains/balance"
	repo "github.com/pavlegich/gophermart/internal/domains/balance/repository"
	"github.com/pavlegich/gophermart/internal/infra/config"
	"github.com/pavlegich/gophermart/internal/infra/logger"
	"github.com/pavlegich/gophermart/internal/utils"
)

type BalanceHandler struct {
	Config  *config.Config
	Service balance.Service
}

type responseBalance struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

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
}

func (h *BalanceHandler) HandleBalanceGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctxValue := ctx.Value(utils.ContextIDKey)
	if ctxValue == nil {
		logger.Log.Info("HandleBalanceGet: get context value failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	userID, ok := ctxValue.(int)
	if !ok {
		logger.Log.Info("HandleBalanceGet: convert context value into integer failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	balanceList, err := h.Service.List(ctx, userID)
	if err != nil {
		logger.Log.Info("HandleBalanceGet: balance get failed")
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
			logger.Log.Info("HandleBalanceGet: action get failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	respJSON, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(respJSON))
}
