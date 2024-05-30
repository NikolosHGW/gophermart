package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/NikolosHGW/gophermart/internal/app/repository"
	"github.com/NikolosHGW/gophermart/internal/domain/entity"
	"go.uber.org/zap"
)

type AccrualService struct {
	repo            repository.AccrualRepository
	logger          *zap.Logger
	accrualAddress  string
	requestInterval time.Duration
}

func NewAccrualService(
	repo repository.AccrualRepository,
	logger *zap.Logger,
	accrualAddress string,
	requestInterval time.Duration,
) *AccrualService {
	return &AccrualService{
		repo:            repo,
		logger:          logger,
		accrualAddress:  accrualAddress,
		requestInterval: requestInterval,
	}
}

func (s *AccrualService) Run(ctx context.Context) {
	ticker := time.NewTicker(s.requestInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.processOrders(ctx)
		}
	}
}

func (s *AccrualService) processOrders(ctx context.Context) {
	orders, err := s.repo.GetNonFinalOrders(ctx)
	if err != nil {
		s.logger.Error("ошибка при получении не обработанных заказов", zap.Error(err))
		return
	}

	var wg sync.WaitGroup
	for _, order := range orders {
		wg.Add(1)
		go func(order entity.Order) {
			defer wg.Done()
			s.processOrder(ctx, order)
		}(order)
	}
	wg.Wait()
}

func (s *AccrualService) processOrder(ctx context.Context, order entity.Order) {
	accrualResponse, err := s.getAccrualData(ctx, order.Number)
	if err != nil {
		s.logger.Error("ошибка при обращении к системе accrual", zap.String("order_number", order.Number), zap.Error(err))
		return
	}

	if accrualResponse != nil {
		err = s.repo.UpdateAccrual(ctx, accrualResponse.Order, accrualResponse.Accrual, accrualResponse.Status)
		if err != nil {
			s.logger.Error("ошибка при обновлении данных accrual", zap.String("order_number", order.Number), zap.Error(err))
		}
	}
}

func (s *AccrualService) getAccrualData(ctx context.Context, orderNumber string) (*AccrualResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.accrualAddress+"/api/orders/"+orderNumber, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании запроса: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка при отправки запроса: %w", err)
	}
	defer func() {
		err := resp.Body.Close()
		s.logger.Error("не удалось закрыть body", zap.Error(err))
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		var accrualResponse AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&accrualResponse); err != nil {
			return nil, fmt.Errorf("ошибка при декодировании ответа: %w", err)
		}
		return &accrualResponse, nil
	case http.StatusNoContent:
		return nil, fmt.Errorf("нет контента")
	case http.StatusTooManyRequests:
		time.Sleep(time.Minute)
		return nil, fmt.Errorf("превышен лимит запросов")
	case http.StatusInternalServerError:
		return nil, fmt.Errorf("ошибка сервера")
	default:
		return nil, fmt.Errorf("неожиданный код состояния: %d", resp.StatusCode)
	}
}

type AccrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}
