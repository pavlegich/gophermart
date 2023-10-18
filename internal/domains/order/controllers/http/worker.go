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

func worker(ctx context.Context, h *OrderHandler) {
	ticker := time.NewTicker(h.Config.Update)

	for {
		select {
		case <-ticker.C:
			jobs, ok := <-h.Jobs
			if !ok {
				logger.Log.Info("worker: channel is closed")
				break
			}

			if h.Config.Accrual == "" {
				logger.Log.Info("worker: accrual address is empty")
				break
			}

			jobOrders := make([]order.Order, 0)

			// Обработка заказов из канала
			for _, ord := range jobs {
				orderNumber := ord.Number

				reqURL := h.Config.Accrual + "/api/orders/" + orderNumber
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
				if err != nil {
					jobOrders = append(jobOrders, ord)
					logger.Log.Info("worker: new request forming failed", zap.Error(err))
					continue
				}

				respJSON, err := utils.GetRequestWithRetry(ctx, req)
				if err != nil {
					jobOrders = append(jobOrders, ord)
					logger.Log.Info("worker: request to accrual system failed",
						zap.String("url", req.RequestURI))
					continue
				}

				// Обработка полученного статуса системы начисления баллов
				if respJSON.StatusCode != http.StatusOK {
					switch respJSON.StatusCode {
					case http.StatusNoContent:
						ord.Status = "PROCESSED"
						ord.Accrual = 0

						if err := h.Service.Upload(ctx, &ord); err != nil {
							if !errors.Is(err, errs.ErrOrderAlreadyProcessed) {
								jobOrders = append(jobOrders, ord)
								logger.Log.Info("worker: upload order failed",
									zap.Error(err))
								continue
							}
						}
					case http.StatusTooManyRequests:
						jobOrders = append(jobOrders, ord)
						retryString := respJSON.Header.Get("Retry-After")
						retry, err := strconv.Atoi(retryString)
						if err != nil {
							jobOrders = append(jobOrders, ord)
							logger.Log.Info("worker: retry header convert failed",
								zap.Error(err),
								zap.String("Retry-After", retryString))
							continue
						}
						logger.Log.Info("worker: status accrual too many requests",
							zap.String("retry-after", retryString))
						time.Sleep(time.Duration(retry) * time.Second)
					case http.StatusInternalServerError:
						jobOrders = append(jobOrders, ord)
						logger.Log.Info("worker: status internal accrual service error")
						continue
					default:
						jobOrders = append(jobOrders, ord)
						logger.Log.Info("worker: unexpected accrual service status code",
							zap.Int("status", respJSON.StatusCode))
						continue
					}
				}

				// Чтение и обработка ответа системы начисления баллов по заказу
				var buf bytes.Buffer
				var resp accrualResponseOrder
				if _, err := buf.ReadFrom(respJSON.Body); err != nil {
					jobOrders = append(jobOrders, ord)
					logger.Log.Info("worker: read response body failed",
						zap.Error(err))
					continue
				}
				if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
					jobOrders = append(jobOrders, ord)
					logger.Log.Info("worker: response unmarshal failed",
						zap.String("body", buf.String()),
						zap.Error(err))
					continue
				}
				respJSON.Body.Close()

				switch resp.Status {
				case "REGISTERED":
					jobOrders = append(jobOrders, ord)
					continue
				case "INVALID":
					ord.Status = resp.Status
					ord.Accrual = 0
				case "PROCESSING":
					ord.Status = resp.Status
					jobOrders = append(jobOrders, ord)
				case "PROCESSED":
					ord.Status = resp.Status
					ord.Accrual = resp.Accrual
				}

				// Загрузка обработанного заказа в хранилище
				if err := h.Service.Upload(ctx, &ord); err != nil {
					if !errors.Is(err, errs.ErrOrderAlreadyProcessed) {
						logger.Log.Info("worker: upload order failed",
							zap.Error(err))
						continue
					}
				}
			}

			// Отправка требующих обработки заказов в канал
			h.Jobs <- jobOrders

		default:
			continue
		}
	}
}
