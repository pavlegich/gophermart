package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/pavlegich/gophermart/internal/domains/order"
	errs "github.com/pavlegich/gophermart/internal/errors"
	"github.com/pavlegich/gophermart/internal/infra/logger"
	"github.com/pavlegich/gophermart/internal/utils"
	"go.uber.org/zap"
)

type accrualResponseOrder struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual,omitempty"`
}

func worker(ctx context.Context, h *OrderHandler, jobs <-chan order.Order) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case ord, ok := <-jobs:
			if !ok {
				return fmt.Errorf("worker: channel is closed")
			}

			orderNumber := ord.Number
			if h.Config.Accrual == "" {
				logger.Log.Info("worker: accrual address is empty")
				continue
			}

			reqURL := h.Config.Accrual + "/api/orders/" + orderNumber
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
			if err != nil {
				logger.Log.Info("worker: new request forming failed", zap.Error(err))
				continue
			}

			respJSON, err := utils.GetRequestWithRetry(ctx, req)
			if err != nil {
				logger.Log.Info("worker: request to accrual system failed",
					zap.String("url", reqURL))
				continue
			}

			var buf bytes.Buffer
			var resp accrualResponseOrder
			if _, err := buf.ReadFrom(respJSON.Body); err != nil {
				logger.Log.Info("worker: read response body failed",
					zap.Error(err))
				continue
			}
			if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
				logger.Log.Info("worker: response unmarshal failed",
					zap.Error(err))
				continue
			}
			respJSON.Body.Close()

			switch resp.Status {
			case "REGISTERED":
				continue
			case "INVALID":
				ord.Status = resp.Status
				ord.Accrual = 0
			case "PROCESSING":
				ord.Status = resp.Status
			case "PROCESSED":
				ord.Status = resp.Status
				ord.Accrual = resp.Accrual
			default:
				logger.Log.Info("worker: invalid response status")
				continue
			}
			if err := h.Service.Upload(ctx, &ord); err != nil {
				if !errors.Is(err, errs.ErrOrderAlreadyProcessed) {
					logger.Log.Info("worker: upload order failed",
						zap.Error(err))
					continue
				}
			}
		}
	}
}
