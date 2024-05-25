package service

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/NikolosHGW/gophermart/internal/app/repository"
	"github.com/NikolosHGW/gophermart/internal/domain/entity"
	"go.uber.org/zap"
)

const (
	initLimit           = 20
	initTimerSeconds    = 60
	initPrevOrderNumber = "0"
)

type Accrual struct {
	prevOrderNumber string
	accrualRepo     repository.AccrualRepository
	logger          *zap.Logger
	accrualAddress  string
	limit           int
}

func NewAccrual(
	accrualRepo repository.AccrualRepository,
	logger *zap.Logger,
	accrualAddress string,
) *Accrual {
	return &Accrual{
		accrualRepo:     accrualRepo,
		logger:          logger,
		accrualAddress:  accrualAddress,
		limit:           initLimit,
		prevOrderNumber: initPrevOrderNumber,
	}
}

func (a *Accrual) StartAccrual() {
	var mutex sync.Mutex

	go func() {
		for {
			mutex.Lock()
			ctx, cancelFunc := context.WithCancel(context.Background())
			a.initRequest(ctx, cancelFunc)
			mutex.Unlock()
		}
	}()
}

func (a *Accrual) initRequest(ctx context.Context, cancelFunc context.CancelFunc) {
	orders, err := a.accrualRepo.GetNonFinalOrders(ctx, a.limit, a.prevOrderNumber)
	if err != nil {
		a.logger.Info("ошибка при получении всех необработанных заказов", zap.Error(err))
	}
	ordersChan := a.generator(ctx, orders)

	for order := range ordersChan {
		select {
		case <-ctx.Done():
			return
		default:
			a.processOrder(ctx, order, cancelFunc)
		}
	}
}

func (a *Accrual) generator(ctx context.Context, orders []entity.Order) chan entity.Order {
	ordersChan := make(chan entity.Order)

	go func() {
		defer close(ordersChan)

		for _, order := range orders {
			select {
			case <-ctx.Done():
				return
			default:
				ordersChan <- order
			}
		}
	}()

	return ordersChan
}

type AccrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

func (a *Accrual) processOrder(ctx context.Context, order entity.Order, cancelFunc context.CancelFunc) {
	accrual, err := a.getAccrualData(order.Number, cancelFunc)
	if err == nil {
		err := a.accrualRepo.UpdateAccrual(ctx, accrual.Order, accrual.Accrual, accrual.Status)
		if err != nil {
			a.logger.Info("ошибка при обновлении статуса и полученных баллов", zap.Error(err))
		}
	}
}

func (a *Accrual) getAccrualData(orderNumber string, cancelFunc context.CancelFunc) (*AccrualResponse, error) {
	resp, err := http.Get("http://" + a.accrualAddress + "/api/orders/" + orderNumber)
	if err != nil {
		a.logger.Info("ошибка при отправке запроса к сервису начисления баллов", zap.Error(err))

		return nil, fmt.Errorf("ошибка при отправке запроса к сервису начисления баллов: %w", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			a.logger.Info("ошибка при закрытии body", zap.Error(err))
		}
	}()

	if resp.StatusCode == http.StatusTooManyRequests {
		cancelFunc()

		newMaxReq, err := extractMaxRequestsFromResponse(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("ошибка при извлечении кол-ва запросов: %w", err)
		}
		a.limit = newMaxReq
		a.prevOrderNumber = orderNumber

		a.logger.Info("превышено число запросов", zap.Error(err))
		time.Sleep(initTimerSeconds * time.Second)
		return nil, fmt.Errorf("превышено число запросов")
	}

	if resp.StatusCode == http.StatusNoContent {
		a.logger.Info("заказ не зарегистрирован в системе расчёта", zap.Error(err))
		return nil, fmt.Errorf("заказ не зарегистрирован в системе расчёта")
	}

	var accrualResponse AccrualResponse
	if err := json.NewDecoder(resp.Body).Decode(&accrualResponse); err != nil {
		a.logger.Info("ошибка при декодирования json", zap.Error(err))

		return nil, fmt.Errorf("ошибка при декодирования json: %w", err)
	}

	return &accrualResponse, nil
}

func extractMaxRequestsFromResponse(body io.ReadCloser) (int, error) {
	scanner := bufio.NewScanner(body)
	re := regexp.MustCompile(`No more than (\d+) requests per minute allowed`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		if len(matches) > 1 {
			maxRequests, err := strconv.Atoi(matches[1])
			if err != nil {
				return 0, fmt.Errorf("не удалось преобразовать строку в число :%w", err)
			}
			return maxRequests, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("ошибка при сканировании тела ответа :%w", err)
	}

	return 0, fmt.Errorf("не найдено строк с ограничением запросов :%w", io.EOF)
}
