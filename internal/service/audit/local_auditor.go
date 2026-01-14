package audit

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/google/uuid"
    "go.uber.org/zap"
)

type localAuditor struct {
    id   string
    file *os.File
    log  *zap.Logger
}

func NewLocalAuditor(fileName string, log *zap.Logger) (Auditor, error) {
    file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0666)
    if err != nil {
        return nil, fmt.Errorf("failed to open file: %w", err)
    }

    return localAuditor{id: uuid.NewString(), file: file, log: log}, nil
}

func (l localAuditor) ID() AuditorID {
    return AuditorID(l.id)
}

func (l localAuditor) Process(payload Payload) {
    if err := json.NewEncoder(l.file).Encode(payload); err != nil {
        l.log.Error("failed to write audit payload to the file", zap.Error(err))
    }
}
