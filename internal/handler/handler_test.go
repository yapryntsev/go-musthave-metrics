package handler

import (
    "fmt"
    "github.com/stretchr/testify/assert"
    "github.com/yapryntsev/go-musthave-metrics/internal/service"
    "github.com/yapryntsev/go-musthave-metrics/internal/service/mocks"
    "log"
    "net/http"
    "net/http/httptest"
    "testing"
)

func Test_GaugeHandler_NonPostMethod_ThrowsMethodNotAllowed(t *testing.T) {
    methods := [...]string{
        http.MethodGet,
        http.MethodHead,
        http.MethodPut,
        http.MethodPatch,
        http.MethodDelete,
        http.MethodConnect,
        http.MethodOptions,
        http.MethodTrace,
    }

    for _, m := range methods {
        t.Run(
            m, func(t *testing.T) {
                // Given
                r := httptest.NewRequest(m, "/gauge/test/100", nil)
                w := httptest.NewRecorder()

                // When
                handler := makeHandler()
                handler.updateGauge(w, r)

                // Then
                res := w.Result()
                defer res.Body.Close()

                assert.Equal(t, res.StatusCode, http.StatusMethodNotAllowed)
            },
        )
    }
}

func Test_CounterHandler_NonPostMethod_ThrowsMethodNotAllowed(t *testing.T) {
    methods := [...]string{
        http.MethodGet,
        http.MethodHead,
        http.MethodPut,
        http.MethodPatch,
        http.MethodDelete,
        http.MethodConnect,
        http.MethodOptions,
        http.MethodTrace,
    }

    for _, m := range methods {
        t.Run(
            m, func(t *testing.T) {
                // Given
                r := httptest.NewRequest(m, "/counter/test/100", nil)
                w := httptest.NewRecorder()

                // When
                handler := makeHandler()
                handler.updateCounter(w, r)

                // Then
                res := w.Result()
                defer res.Body.Close()

                assert.Equal(t, res.StatusCode, http.StatusMethodNotAllowed)
            },
        )
    }
}

func Test_GaugeHandler_InvalidPath_ThrowsNotFound(t *testing.T) {
    tests := [...]struct {
        name  string
        value string
    }{
        {"", ""},
        {"", "100.2"},
        {"test-name", ""},
    }

    for _, c := range tests {
        t.Run(
            c.name, func(t *testing.T) {
                // Given
                r := httptest.NewRequest(
                    http.MethodPost,
                    fmt.Sprintf(`/gauge/%s/%s`, c.name, c.value),
                    nil,
                )
                w := httptest.NewRecorder()

                r.SetPathValue(service.MetricNamePathKey, c.name)
                r.SetPathValue(service.MetricValuePathKey, c.value)

                // When
                handler := makeHandler()
                handler.updateGauge(w, r)

                // Then
                res := w.Result()
                defer res.Body.Close()

                assert.Equal(t, res.StatusCode, http.StatusNotFound)
            },
        )
    }
}

func Test_CounterHandler_InvalidPath_ThrowsNotFound(t *testing.T) {
    tests := [...]struct {
        name  string
        value string
    }{
        {"", ""},
        {"", "100"},
        {"test-name", ""},
    }

    for _, c := range tests {
        t.Run(
            c.name, func(t *testing.T) {
                // Given
                r := httptest.NewRequest(
                    http.MethodPost,
                    fmt.Sprintf(`/counter/%s/%s`, c.name, c.value),
                    nil,
                )
                w := httptest.NewRecorder()

                r.SetPathValue(service.MetricNamePathKey, c.name)
                r.SetPathValue(service.MetricValuePathKey, c.value)

                // When
                handler := makeHandler()
                handler.updateCounter(w, r)

                // Then
                res := w.Result()
                defer res.Body.Close()

                assert.Equal(t, res.StatusCode, http.StatusNotFound)
            },
        )
    }
}

func Test_GaugeHandler_NonFloatValue_ThrowsBadRequest(t *testing.T) {
    // Given
    r := httptest.NewRequest(http.MethodPost, "/gauge/test/value", nil)
    w := httptest.NewRecorder()

    r.SetPathValue(service.MetricNamePathKey, "test")
    r.SetPathValue(service.MetricValuePathKey, "value")

    // When
    handler := makeHandler()
    handler.updateGauge(w, r)

    // Then
    res := w.Result()
    defer res.Body.Close()

    assert.Equal(t, res.StatusCode, http.StatusBadRequest)
}

func Test_CounterHandler_NonIntValue_ThrowsBadRequest(t *testing.T) {
    // Given
    r := httptest.NewRequest(http.MethodPost, "/counter/test/value", nil)
    w := httptest.NewRecorder()

    r.SetPathValue(service.MetricNamePathKey, "test")
    r.SetPathValue(service.MetricValuePathKey, "value")

    // When
    handler := makeHandler()
    handler.updateGauge(w, r)

    // Then
    res := w.Result()
    defer res.Body.Close()

    assert.Equal(t, res.StatusCode, http.StatusBadRequest)
}

func makeHandler() *MetricHandler {
    return &MetricHandler{
        log:     log.Default(),
        service: &mocks.MetricServiceMock{},
    }
}
