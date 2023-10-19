package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

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

// workerCheckOrders получает и отправляет в канал необработанные заказы
func workerCheckOrders(ctx context.Context, h *OrderHandler) {
	ticker := time.NewTicker(h.Config.Update)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Получение списка заказов
			ordersList, err := h.Service.ListUnprocessed(ctx)
			if err != nil {
				if errors.Is(err, errs.ErrOrdersNotFound) {
					logger.Log.Info("workerCheckOrders: orders not found for this user",
						zap.Error(err))
				} else {
					logger.Log.Info("workerCheckOrders: get orders list failed",
						zap.Error(err))
				}
				continue
			}

			// Отправка всех необработанных заказов в канал
			for _, o := range ordersList {
				job := order.Order{
					ID:        o.ID,
					Number:    o.Number,
					UserID:    o.UserID,
					Status:    o.Status,
					Accrual:   o.Accrual,
					CreatedAt: o.CreatedAt,
				}
				h.Jobs <- job
			}
		}
	}
}

// workerRequestAccrual получает и обрабатывает ответ от системы начисления баллов по заказам
func workerRequestAccrual(ctx context.Context, h *OrderHandler, jobs <-chan order.Order) {
	for {
		select {
		case <-ctx.Done():
			return
		case ord, ok := <-jobs:
			if !ok {
				logger.Log.Info("workerRequestAccrual: channel is closed")
				return
			}

			orderNumber := ord.Number
			if h.Config.Accrual == "" {
				logger.Log.Info("workerRequestAccrual: accrual address is empty")
				return
			}

			reqURL := h.Config.Accrual + "/api/orders/" + orderNumber
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
			if err != nil {
				logger.Log.Info("workerRequestAccrual: new request forming failed", zap.Error(err))
				continue
			}

			// Запрос только в случае, если таймер остановлен
			<-h.RequestTimer.C
			resp, err := utils.GetRequestWithRetry(ctx, req)
			if err != nil {
				logger.Log.Info("workerRequestAccrual: request to accrual system failed",
					zap.String("url", req.RequestURI))
				continue
			}
			defer resp.Body.Close()

			// Обработка полученного статуса системы начисления баллов
			if resp.StatusCode != http.StatusOK {
				switch resp.StatusCode {
				case http.StatusNoContent:
					ord.Status = "INVALID"
					ord.Accrual = 0
				case http.StatusTooManyRequests:
					retryString := resp.Header.Get("Retry-After")
					retry, err := strconv.Atoi(retryString)
					if err != nil {
						logger.Log.Info("workerRequestAccrual: retry header convert into integer failed",
							zap.Error(err),
							zap.String("Retry-After", retryString))
						continue
					}
					logger.Log.Info("workerRequestAccrual: status accrual too many requests",
						zap.String("retry-after", retryString))
					h.RequestTimer.Reset(time.Duration(retry) * time.Second)
					continue
				case http.StatusInternalServerError:
					logger.Log.Info("workerRequestAccrual: status internal accrual service error")
					continue
				default:
					logger.Log.Info("workerRequestAccrual: unexpected accrual service status code",
						zap.Int("status", resp.StatusCode))
					continue
				}
			}

			// Обработка тела ответа системы начисления баллов
			var buf bytes.Buffer
			var respJSON accrualResponseOrder
			if _, err := buf.ReadFrom(resp.Body); err != nil {
				logger.Log.Info("workerRequestAccrual: read response body failed",
					zap.Error(err))
				continue
			}
			if err := json.Unmarshal(buf.Bytes(), &respJSON); err != nil {
				logger.Log.Info("workerRequestAccrual: response unmarshal failed",
					zap.String("body", buf.String()),
					zap.Error(err))
				continue
			}

			// Проверка статуса обработки заказа в системе начисления баллов
			if resp.StatusCode == http.StatusOK {
				switch respJSON.Status {
				case "REGISTERED":
					continue
				case "INVALID":
					ord.Status = respJSON.Status
					ord.Accrual = 0
				case "PROCESSING":
					ord.Status = respJSON.Status
				case "PROCESSED":
					ord.Status = respJSON.Status
					ord.Accrual = respJSON.Accrual
				default:
					logger.Log.Info("workerRequestAccrual: invalid response order status",
						zap.String("status", respJSON.Status))
					continue
				}
			}

			// Загрузка обновленного заказа в хранилище
			if err := h.Service.Upload(ctx, &ord); err != nil {
				if !errors.Is(err, errs.ErrOrderAlreadyProcessed) {
					logger.Log.Info("workerRequestAccrual: upload order failed",
						zap.Error(err))
				}
			}
		}
	}
}
