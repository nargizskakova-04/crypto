package concurrency

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"crypto/internal/core/domain"
	"crypto/internal/core/port"
)

// DataProcessor обрабатывает входящие данные и агрегирует их
type DataProcessor struct {
	batchWriter         *BatchWriter
	cacheRepo           port.CacheRepository
	aggregationInterval time.Duration
	workersPerExchange  int

	priceHistory map[string][]domain.PriceUpdate
	historyMutex sync.RWMutex

	inputCh           chan domain.PriceUpdate
	workerPool        *WorkerPool
	aggregationTicker *time.Ticker
	stopCh            chan struct{}
	wg                sync.WaitGroup
}

func NewDataProcessor(
	batchWriter *BatchWriter,
	cacheRepo port.CacheRepository,
	aggregationInterval time.Duration,
	workersPerExchange int,
) *DataProcessor {
	return &DataProcessor{
		batchWriter:         batchWriter,
		cacheRepo:           cacheRepo,
		aggregationInterval: aggregationInterval,
		workersPerExchange:  workersPerExchange,
		priceHistory:        make(map[string][]domain.PriceUpdate),
		inputCh:             make(chan domain.PriceUpdate, 1000),
		aggregationTicker:   time.NewTicker(aggregationInterval),
		stopCh:              make(chan struct{}),
	}
}

func (dp *DataProcessor) Start(ctx context.Context) error {
	slog.Info("Starting DataProcessor", "workers", dp.workersPerExchange*3, "aggregation_interval", dp.aggregationInterval)

	// Создаем worker pool
	dp.workerPool = NewWorkerPool(dp.workersPerExchange*3, 1000) // 3 биржи * воркеры на биржу
	dp.workerPool.Start()

	// Запускаем основной цикл
	dp.wg.Add(1)
	go dp.run(ctx)

	return nil
}

func (dp *DataProcessor) run(ctx context.Context) {
	defer dp.wg.Done()
	defer dp.aggregationTicker.Stop()

	for {
		select {
		case update := <-dp.inputCh:
			// Отправляем в worker pool для обработки
			task := WorkerTask{
				Data: update,
				ProcessFunc: func(data interface{}) error {
					return dp.processSingleUpdate(data.(domain.PriceUpdate))
				},
			}

			if err := dp.workerPool.SubmitTask(task); err != nil {
				slog.Error("Failed to submit task to worker pool", "error", err)
			}

		case <-dp.aggregationTicker.C:
			dp.performAggregation()

		case <-ctx.Done():
			slog.Info("DataProcessor context cancelled")
			dp.performAggregation() // Последняя агрегация
			return

		case <-dp.stopCh:
			slog.Info("DataProcessor stop signal received")
			dp.performAggregation() // Последняя агрегация
			return
		}
	}
}

func (dp *DataProcessor) ProcessPriceUpdate(update domain.PriceUpdate) error {
	select {
	case dp.inputCh <- update:
		return nil
	default:
		return fmt.Errorf("data processor input buffer is full")
	}
}

func (dp *DataProcessor) processSingleUpdate(update domain.PriceUpdate) error {
	// Валидируем данные
	if err := update.Validate(); err != nil {
		return fmt.Errorf("invalid price update: %w", err)
	}

	// Сохраняем в Redis (если доступен)
	if dp.cacheRepo != nil {
		if err := dp.cacheRepo.SetLatestPrice(context.Background(), update); err != nil {
			slog.Error("Failed to save to Redis", "error", err)
			// Не прерываем обработку, продолжаем без Redis
		}
	}

	// Сохраняем в истории для агрегации
	dp.historyMutex.Lock()
	key := update.GetKey()
	dp.priceHistory[key] = append(dp.priceHistory[key], update)

	// Очищаем старые данные (старше 2 минут)
	cutoff := time.Now().Add(-2 * time.Minute)
	filtered := make([]domain.PriceUpdate, 0)
	for _, price := range dp.priceHistory[key] {
		if price.Timestamp.After(cutoff) {
			filtered = append(filtered, price)
		}
	}
	dp.priceHistory[key] = filtered
	dp.historyMutex.Unlock()

	slog.Debug("Processed price update", "exchange", update.Exchange, "symbol", update.Symbol, "price", update.Price)
	return nil
}

func (dp *DataProcessor) performAggregation() {
	dp.historyMutex.RLock()
	defer dp.historyMutex.RUnlock()

	now := time.Now()
	cutoff := now.Add(-1 * time.Minute) // Данные за последнюю минуту
	aggregations := make([]domain.AggregatedData, 0)

	slog.Info("Performing minute aggregation", "timestamp", now.Format("15:04:05"))

	for key, prices := range dp.priceHistory {
		if len(prices) == 0 {
			continue
		}

		// Фильтруем цены за последнюю минуту
		recentPrices := make([]domain.PriceUpdate, 0)
		for _, price := range prices {
			if price.Timestamp.After(cutoff) {
				recentPrices = append(recentPrices, price)
			}
		}

		if len(recentPrices) == 0 {
			continue
		}

		// Вычисляем статистику
		sum := 0.0
		minPrice := recentPrices[0].Price
		maxPrice := recentPrices[0].Price

		for _, price := range recentPrices {
			sum += price.Price
			if price.Price < minPrice {
				minPrice = price.Price
			}
			if price.Price > maxPrice {
				maxPrice = price.Price
			}
		}

		avgPrice := sum / float64(len(recentPrices))

		aggregation := domain.AggregatedData{
			PairName:     recentPrices[0].Symbol,
			Exchange:     recentPrices[0].Exchange,
			Timestamp:    now,
			AveragePrice: avgPrice,
			MinPrice:     minPrice,
			MaxPrice:     maxPrice,
		}

		aggregations = append(aggregations, aggregation)

		slog.Info("Aggregated data",
			"key", key,
			"updates", len(recentPrices),
			"avg", avgPrice,
			"min", minPrice,
			"max", maxPrice)
	}

	// Отправляем агрегации в BatchWriter
	for _, agg := range aggregations {
		if err := dp.batchWriter.AddData(agg); err != nil {
			slog.Error("Failed to add aggregation to batch writer", "error", err)
		}
	}

	slog.Info("Minute aggregation completed", "aggregations_created", len(aggregations))
}

func (dp *DataProcessor) Stop() error {
	close(dp.stopCh)
	if dp.workerPool != nil {
		dp.workerPool.Stop()
	}
	dp.wg.Wait()
	return nil
}

// pkg/concurrency/worker_pool.go

// WorkerPool для обработки задач
type WorkerPool struct {
	workerCount int
	taskQueue   chan WorkerTask
	wg          sync.WaitGroup
	stopCh      chan struct{}
}

type WorkerTask struct {
	Data        interface{}
	ProcessFunc func(interface{}) error
}

func NewWorkerPool(workerCount, queueSize int) *WorkerPool {
	return &WorkerPool{
		workerCount: workerCount,
		taskQueue:   make(chan WorkerTask, queueSize),
		stopCh:      make(chan struct{}),
	}
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i + 1)
	}
}

func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	for {
		select {
		case task := <-wp.taskQueue:
			if err := task.ProcessFunc(task.Data); err != nil {
				slog.Error("Worker task failed", "worker_id", id, "error", err)
			}

		case <-wp.stopCh:
			slog.Info("Worker stopping", "worker_id", id)
			return
		}
	}
}

func (wp *WorkerPool) SubmitTask(task WorkerTask) error {
	select {
	case wp.taskQueue <- task:
		return nil
	default:
		return fmt.Errorf("worker pool task queue is full")
	}
}

func (wp *WorkerPool) Stop() {
	close(wp.stopCh)
	wp.wg.Wait()
}

// pkg/concurrency/fan_in.go

// FanIn объединяет несколько каналов в один (для демонстрации паттерна)
func FanIn(inputs ...<-chan domain.PriceUpdate) <-chan domain.PriceUpdate {
	outputCh := make(chan domain.PriceUpdate)
	var wg sync.WaitGroup

	// Для каждого входного канала запускаем горутину
	for _, input := range inputs {
		wg.Add(1)
		go func(ch <-chan domain.PriceUpdate) {
			defer wg.Done()

			// Перенаправляем все данные из input в output
			for update := range ch {
				outputCh <- update
			}
		}(input)
	}

	// Закрываем выходной канал когда все входные закончились
	go func() {
		wg.Wait()
		close(outputCh)
	}()

	return outputCh
}

// pkg/concurrency/fan_out.go

// FanOut распределяет данные от одного источника по нескольким воркерам
func FanOut(inputCh <-chan domain.PriceUpdate, workerCount int, processorFunc func(domain.PriceUpdate) error) {
	var wg sync.WaitGroup

	// Запускаем воркеров
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for update := range inputCh {
				if err := processorFunc(update); err != nil {
					slog.Error("Fan-out worker error", "worker_id", workerID, "error", err)
				}
			}
		}(i + 1)
	}

	// Ждем завершения всех воркеров
	wg.Wait()
}
