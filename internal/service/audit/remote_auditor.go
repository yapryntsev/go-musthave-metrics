package audit

import (
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type remoteAuditor struct {
	id     string
	url    string
	client *resty.Client
	log    *zap.Logger
}

// NewRemoteAuditor returns an Auditor instance that stores data on a remote server.
func NewRemoteAuditor(url string, client *resty.Client, log *zap.Logger) Auditor {
	return remoteAuditor{id: uuid.NewString(), url: url, client: client, log: log}
}

func (l remoteAuditor) ID() AuditorID {
	return AuditorID(l.id)
}

func (l remoteAuditor) Process(payload Payload) {
	resp, err := l.client.R().
		SetBody(payload).
		Post(l.url)

	if err != nil {
		l.log.Error("failed to send metrics", zap.Error(err))
	}

	if resp == nil {
		l.log.Error("failed to get response")
		return
	}

	if resp.StatusCode() != http.StatusOK {
		l.log.Error("unexpected status code", zap.Int("code", resp.StatusCode()))
	}
}
