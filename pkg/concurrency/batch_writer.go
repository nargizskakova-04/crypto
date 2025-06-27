// pkg/concurrency/batch_writer.go
package concurrency

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"crypto/internal/core/domain"
)

// BatchWriter группирует данные и записывает их батчами в PostgreSQL
type BatchWriter struct {
	db           *sql.DB
	batchSize    int
	batchTimeout time.Duration

	buffer      []domain.AggregatedData
	bufferMutex sync.Mutex

	inputCh chan domain.AggregatedData
	ticker  *time.Ticker
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

func NewBatchWriter(db *sql.DB, batchSize int, batchTimeout time.Duration) *BatchWriter {
	return &BatchWriter{
		db:           db,
		batchSize:    batchSize,
		batchTimeout: batchTimeout,
		buffer:       make([]domain.AggregatedData, 0, batchSize),
		inputCh:      make(chan domain.AggregatedData, batchSize*2),
		ticker:       time.NewTicker(batchTimeout),
		stopCh:       make(chan struct{}),
	}
}

func (bw *BatchWriter) Start(ctx context.Context) error {
	slog.Info("Starting BatchWriter", "batch_size", bw.batchSize, "timeout", bw.batchTimeout)

	bw.wg.Add(1)
	go bw.run(ctx)

	return nil
}

func (bw *BatchWriter) run(ctx context.Context) {
	defer bw.wg.Done()
	defer bw.ticker.Stop()

	for {
		select {
		case data := <-bw.inputCh:
			bw.addToBuffer(data)

		case <-bw.ticker.C:
			bw.flushBuffer()

		case <-ctx.Done():
			slog.Info("BatchWriter context cancelled, flushing remaining data")
			bw.flushBuffer()
			return

		case <-bw.stopCh:
			slog.Info("BatchWriter stop signal received, flushing remaining data")
			bw.flushBuffer()
			return
		}
	}
}

func (bw *BatchWriter) AddData(data domain.AggregatedData) error {
	select {
	case bw.inputCh <- data:
		return nil
	default:
		return fmt.Errorf("batch writer buffer is full")
	}
}

func (bw *BatchWriter) addToBuffer(data domain.AggregatedData) {
	bw.bufferMutex.Lock()
	defer bw.bufferMutex.Unlock()

	bw.buffer = append(bw.buffer, data)

	if len(bw.buffer) >= bw.batchSize {
		go bw.flushBuffer() // Асинхронная запись
	}
}

func (bw *BatchWriter) flushBuffer() {
	bw.bufferMutex.Lock()
	defer bw.bufferMutex.Unlock()

	if len(bw.buffer) == 0 {
		return
	}

	slog.Info("Flushing batch to PostgreSQL", "count", len(bw.buffer))

	if err := bw.writeBatch(bw.buffer); err != nil {
		slog.Error("Failed to write batch to PostgreSQL", "error", err, "count", len(bw.buffer))
		return
	}

	slog.Info("Successfully wrote batch to PostgreSQL", "count", len(bw.buffer))
	bw.buffer = bw.buffer[:0] // Очищаем буфер
}

func (bw *BatchWriter) writeBatch(data []domain.AggregatedData) error {
	if len(data) == 0 {
		return nil
	}

	// Строим bulk insert запрос
	query := `INSERT INTO prices (pair_name, exchange, timestamp, average_price, min_price, max_price) VALUES `

	values := make([]string, len(data))
	args := make([]interface{}, 0, len(data)*6)

	for i, item := range data {
		values[i] = fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)",
			i*6+1, i*6+2, i*6+3, i*6+4, i*6+5, i*6+6)

		args = append(args,
			item.PairName,
			item.Exchange,
			item.Timestamp,
			item.AveragePrice,
			item.MinPrice,
			item.MaxPrice)
	}

	query += strings.Join(values, ", ")

	// Выполняем в транзакции
	tx, err := bw.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute batch insert: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (bw *BatchWriter) Stop() error {
	close(bw.stopCh)
	bw.wg.Wait()
	return nil
}
