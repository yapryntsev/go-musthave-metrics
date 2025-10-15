package middleware

import (
    log "github.com/sirupsen/logrus"
    logTest "github.com/sirupsen/logrus/hooks/test"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "net/http"
    "net/http/httptest"
    "testing"
)

func Test_MiddlewareProduceLogEntry(t *testing.T) {
    // Given
    expectedURI := "/test"
    expectedMethod := http.MethodPost
    expectedStatus := http.StatusMethodNotAllowed
    expectedResponse := []byte("test body")

    logger, hook := logTest.NewNullLogger()

    r := httptest.NewRequest(expectedMethod, expectedURI, nil)
    w := httptest.NewRecorder()

    handler := func(w http.ResponseWriter, r *http.Request) {
        _, _ = w.Write(expectedResponse)
        w.WriteHeader(expectedStatus)
    }
    middleware := Logger(log.NewEntry(logger))

    // When
    middleware(http.HandlerFunc(handler)).ServeHTTP(w, r)

    // Then
    res := w.Result()
    defer res.Body.Close()

    require.NotEmpty(t, hook.Entries, "expected at least one log entry")
    entry := hook.LastEntry()

    assert.Equal(t, log.InfoLevel, entry.Level)

    assert.Contains(t, entry.Data, "uri")
    assert.Contains(t, entry.Data, "method")
    assert.Contains(t, entry.Data, "duration")
    assert.Contains(t, entry.Data, "status")
    assert.Contains(t, entry.Data, "size")

    assert.Equal(t, entry.Data["uri"], "/test")
    assert.Equal(t, entry.Data["method"], expectedMethod)
    assert.Equal(t, entry.Data["status"], expectedStatus)
    assert.Equal(t, entry.Data["size"], len(expectedResponse))
}

func Test_HandlerNotCallWriteHeader_TreatAsStatusOK(t *testing.T) {
    // Given
    logger, hook := logTest.NewNullLogger()

    r := httptest.NewRequest(http.MethodPost, "/test", nil)
    w := httptest.NewRecorder()

    handler := func(w http.ResponseWriter, r *http.Request) {}
    middleware := Logger(log.NewEntry(logger))

    // When
    middleware(http.HandlerFunc(handler)).ServeHTTP(w, r)

    // Then
    res := w.Result()
    defer res.Body.Close()

    require.NotEmpty(t, hook.Entries, "expected at least one log entry")
    entry := hook.LastEntry()

    assert.Contains(t, entry.Data, "status")
    assert.Equal(t, entry.Data["status"], http.StatusOK)
}
