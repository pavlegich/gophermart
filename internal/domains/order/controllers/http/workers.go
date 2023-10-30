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
				if !errors.Is(err, errs.ErrOrdersNotFound) {
					logger.Log.Error("workerCheckOrders: get orders list failed",
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
	timer := time.NewTimer(0)

	for {
		select {
		case <-ctx.Done():
			return
		case ord, ok := <-jobs:
			if !ok {
				logger.Log.Error("workerRequestAccrual: channel is closed")
				return
			}

			orderNumber := ord.Number
			if h.Config.Accrual == "" {
				logger.Log.With(zap.String("order_id", orderNumber)).Error("workerRequestAccrual: accrual address is empty")
				return
			}

			reqURL := h.Config.Accrual + "/api/orders/" + orderNumber
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
			if err != nil {
				logger.Log.With(zap.String("order_id", orderNumber)).Error("workerRequestAccrual: new request forming failed", zap.Error(err))
				continue
			}

			// Запрос только в случае, если таймер остановлен
			<-timer.C
			resp, err := utils.GetRequestWithRetry(ctx, req)
			if err != nil {
				logger.Log.With(zap.String("order_id", orderNumber)).Error("workerRequestAccrual: request to accrual system failed",
					zap.String("url", reqURL), zap.Error(err))
				continue
			}
			defer resp.Body.Close()

			// Обработка полученного статуса системы начисления баллов
			if resp.StatusCode != http.StatusOK {
				switch resp.StatusCode {
				case http.StatusNoContent:
					logger.Log.With(zap.String("order_id", orderNumber)).Error("workerRequestAccrual: order not found in accrual service")
				case http.StatusTooManyRequests:
					retryString := resp.Header.Get("Retry-After")
					retry, err := strconv.Atoi(retryString)
					if err != nil {
						logger.Log.With(zap.String("order_id", orderNumber)).Error("workerRequestAccrual: retry header convert into integer failed",
							zap.Error(err),
							zap.String("Retry-After", retryString))
					}
					logger.Log.With(zap.String("order_id", orderNumber)).Error("workerRequestAccrual: status accrual too many requests",
						zap.String("retry-after", retryString))
					timer.Reset(time.Duration(retry) * time.Second)
				case http.StatusInternalServerError:
					logger.Log.With(zap.String("order_id", orderNumber)).Error("workerRequestAccrual: status internal accrual service error")
				default:
					logger.Log.With(zap.String("order_id", orderNumber)).Error("workerRequestAccrual: unexpected accrual service status code",
						zap.Int("status", resp.StatusCode))
				}
				continue
			}

			// Обработка тела ответа системы начисления баллов
			var buf bytes.Buffer
			var respJSON accrualResponseOrder
			if _, err := buf.ReadFrom(resp.Body); err != nil {
				logger.Log.With(zap.String("order_id", orderNumber)).Error("workerRequestAccrual: read response body failed",
					zap.Error(err))
				continue
			}
			if err := json.Unmarshal(buf.Bytes(), &respJSON); err != nil {
				logger.Log.With(zap.String("order_id", orderNumber)).Error("workerRequestAccrual: response unmarshal failed",
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
					logger.Log.With(zap.String("order_id", orderNumber)).Error("workerRequestAccrual: invalid response order status",
						zap.String("status", respJSON.Status))
					continue
				}
			}

			// Загрузка обновленного заказа в хранилище
			if err := h.Service.Upload(ctx, &ord); err != nil {
				logger.Log.With(zap.String("order_id", orderNumber)).Error("workerRequestAccrual: upload order failed",
					zap.Error(err))
			}
		}
	}
}
