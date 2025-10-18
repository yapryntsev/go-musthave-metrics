package repository

import (
    "bytes"
    "encoding/json"
    "errors"
    log "github.com/sirupsen/logrus"
    models "github.com/yapryntsev/go-musthave-metrics/internal/model"
    "os"
    "time"
)

type FileMetricRepository struct {
    log      *log.Entry
    inMemory *MemoryMetricRepository

    filePath   string
    storeInt   time.Duration
    lastUpdate time.Time
}

func NewFileRepo(storeInt time.Duration, filePath string, restore bool, log *log.Entry) (*FileMetricRepository, error) {
    repo := &FileMetricRepository{
        log:        log,
        inMemory:   NewInMemoryRepo(),
        filePath:   filePath,
        storeInt:   storeInt,
        lastUpdate: time.Now(),
    }

    if restore {
        repo.ReadFromFile()
    }

    return repo, nil
}

func (r *FileMetricRepository) GetAll() ([]*models.Metrics, error) {
    return r.inMemory.GetAll()
}

func (r *FileMetricRepository) Get(mID string, mType string) (*models.Metrics, error) {
    return r.inMemory.Get(mID, mType)
}

func (r *FileMetricRepository) Set(metric *models.Metrics) error {
    err := r.inMemory.Set(metric)
    if err != nil {
        return err
    }

    if r.storeInt == 0 || time.Since(r.lastUpdate) >= r.storeInt {
        r.WriteToFile()
    }

    return nil
}

func (r *FileMetricRepository) WriteToFile() {
    var buf bytes.Buffer

    r.inMemory.mu.Lock()
    defer r.inMemory.mu.Unlock()

    if err := json.NewEncoder(&buf).Encode(r.inMemory.storage); err != nil {
        r.log.WithField(
            "err", err,
        ).Fatal("failed to save repo state to file")
    }

    if err := os.WriteFile(r.filePath, buf.Bytes(), 0644); err != nil {
        r.log.WithField(
            "err", err,
        ).Fatal("failed to save repo state to file")
    }

    r.lastUpdate = time.Now()
}

func (r *FileMetricRepository) ReadFromFile() {
    r.log.Info("restoring repo state from file")

    b, err := os.ReadFile(r.filePath)
    if err != nil {
        if errors.Is(err, os.ErrNotExist) {
            r.log.Info("file does not exist")
            return
        }

        r.log.WithField(
            "err", err,
        ).Fatal("failed to restore repo state from file")
    }

    r.inMemory.mu.Lock()
    defer r.inMemory.mu.Unlock()

    if err := json.NewDecoder(bytes.NewReader(b)).Decode(&r.inMemory.storage); err != nil {
        r.log.WithField(
            "err", err,
        ).Fatal("failed to restore repo state from file")
    }
}
