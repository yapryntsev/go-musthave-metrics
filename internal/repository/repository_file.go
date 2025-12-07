package repository

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    models "github.com/yapryntsev/go-musthave-metrics/internal/model"
    "go.uber.org/zap"
    "os"
    "time"
)

type FileMetricRepository struct {
    log      *zap.Logger
    inMemory *MemoryMetricRepository

    filePath   string
    storeInt   time.Duration
    lastUpdate time.Time
}

func newFileRepository(storeInt time.Duration, filePath string, restore bool, log *zap.Logger) *FileMetricRepository {
    repo := &FileMetricRepository{
        log:        log,
        inMemory:   newInMemoryRepo(),
        filePath:   filePath,
        storeInt:   storeInt,
        lastUpdate: time.Now(),
    }

    if restore {
        repo.ReadFromFile()
    }

    return repo
}

func (r *FileMetricRepository) GetAll(ctx context.Context) ([]models.Metrics, error) {
    return r.inMemory.GetAll(ctx)
}

func (r *FileMetricRepository) Get(ctx context.Context, mID string, mType string) (*models.Metrics, error) {
    return r.inMemory.Get(ctx, mID, mType)
}

func (r *FileMetricRepository) Set(ctx context.Context, metric models.Metrics) error {
    err := r.inMemory.Set(ctx, metric)
    if err != nil {
        return err
    }

    if r.storeInt == 0 || time.Since(r.lastUpdate) >= r.storeInt {
        r.WriteToFile()
    }

    return nil
}

func (r *FileMetricRepository) SetBatch(ctx context.Context, metrics []models.Metrics) error {
    for _, m := range metrics {
        err := r.inMemory.Set(ctx, m)
        if err != nil {
            return err
        }
    }

    if r.storeInt == 0 || time.Since(r.lastUpdate) >= r.storeInt {
        r.WriteToFile()
    }

    return nil
}

func (r *FileMetricRepository) WriteToFile() {
    var buf bytes.Buffer

    r.inMemory.mu.Lock()

    if err := json.NewEncoder(&buf).Encode(r.inMemory.storage); err != nil {
        r.log.Fatal("failed to encode in memory storage", zap.Error(err))
    }

    r.inMemory.mu.Unlock()

    if err := os.WriteFile(r.filePath, buf.Bytes(), 0644); err != nil {
        r.log.Fatal(fmt.Sprintf("failed to save repo state to file %s", r.filePath), zap.Error(err))
    }

    r.lastUpdate = time.Now()
}

func (r *FileMetricRepository) ReadFromFile() {
    r.log.Debug("restoring repo state from file")

    b, err := os.ReadFile(r.filePath)
    if err != nil {
        if errors.Is(err, os.ErrNotExist) {
            r.log.Debug(fmt.Sprintf("file %s does not exists", r.filePath))
            return
        }

        r.log.Fatal(fmt.Sprintf("failed to restore repo state from file %s", r.filePath), zap.Error(err))
    }

    r.inMemory.mu.Lock()
    defer r.inMemory.mu.Unlock()

    if err := json.NewDecoder(bytes.NewReader(b)).Decode(&r.inMemory.storage); err != nil {
        r.log.Fatal(fmt.Sprintf("failed to decode repo state from file %s", r.filePath), zap.Error(err))
    }
}
