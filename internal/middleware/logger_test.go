package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func Test_MiddlewareProduceLogEntry(t *testing.T) {
	// Given
	expectedURI := "/test"
	expectedMethod := http.MethodPost
	expectedStatus := http.StatusMethodNotAllowed
	expectedResponse := []byte("test body")

	observed, logs := observer.New(zap.DebugLevel)
	logger := zap.New(observed)

	r := httptest.NewRequest(expectedMethod, expectedURI, nil)
	w := httptest.NewRecorder()

	handler := func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(expectedResponse)
		w.WriteHeader(expectedStatus)
	}
	middleware := Logger(logger)

	// When
	middleware(http.HandlerFunc(handler)).ServeHTTP(w, r)

	// Then
	res := w.Result()
	defer res.Body.Close()

	require.NotEmpty(t, logs, "expected at least one log entry")
	entry := logs.All()[0]
	context := entry.ContextMap()

	assert.Equal(t, zap.DebugLevel, entry.Level)

	assert.Contains(t, context, "uri")
	assert.Contains(t, context, "method")
	assert.Contains(t, context, "duration")
	assert.Contains(t, context, "status")
	assert.Contains(t, context, "size")

	assert.Equal(t, "/test", context["uri"])
	assert.Equal(t, expectedMethod, context["method"])
	assert.Equal(t, int64(expectedStatus), context["status"])
	assert.Equal(t, int64(len(expectedResponse)), context["size"])
}

func Test_HandlerNotCallWriteHeader_TreatAsStatusOK(t *testing.T) {
	// Given
	observed, logs := observer.New(zap.DebugLevel)
	logger := zap.New(observed)

	r := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()

	handler := func(w http.ResponseWriter, r *http.Request) {}
	middleware := Logger(logger)

	// When
	middleware(http.HandlerFunc(handler)).ServeHTTP(w, r)

	// Then
	res := w.Result()
	defer res.Body.Close()

	require.NotEmpty(t, logs, "expected at least one log entry")
	entry := logs.All()[0]
	context := entry.ContextMap()

	assert.Contains(t, context, "status")
	assert.Equal(t, int64(http.StatusOK), context["status"])
}
