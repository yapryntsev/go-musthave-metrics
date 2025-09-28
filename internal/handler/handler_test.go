package handler

import (
    "errors"
    "fmt"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/yapryntsev/go-musthave-metrics/internal/service"
    "github.com/yapryntsev/go-musthave-metrics/internal/service/mocks"
    "io"
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
                handler := makeHandler(mocks.NewServiceMock())
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
                handler := makeHandler(mocks.NewServiceMock())
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
                handler := makeHandler(mocks.NewServiceMock())
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
                handler := makeHandler(mocks.NewServiceMock())
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
    handler := makeHandler(mocks.NewServiceMock())
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
    handler := makeHandler(mocks.NewServiceMock())
    handler.updateGauge(w, r)

    // Then
    res := w.Result()
    defer res.Body.Close()

    assert.Equal(t, res.StatusCode, http.StatusBadRequest)
}

func makeHandler(service *mocks.MetricServiceMock) *MetricHandler {
    return &MetricHandler{
        log:     log.Default(),
        service: service,
    }
}

func Test_GetAllHandler_HasValue_Return(t *testing.T) {
    // Given
    expectedName := `test`
    expectedValue := `16`

    r := httptest.NewRequest(http.MethodGet, "/", nil)
    w := httptest.NewRecorder()

    service := mocks.NewServiceMock()
    handler := makeHandler(service)

    service.GetAllReturnValue = map[string]string{
        expectedName: expectedValue,
    }

    // When
    handler.GetAll(w, r)

    // Then
    res := w.Result()
    defer res.Body.Close()

    body, err := io.ReadAll(res.Body)
    require.NoError(t, err, `failed to read response body`)

    require.Equal(t, res.StatusCode, http.StatusOK)
    require.Equal(t, body, []byte(fmt.Sprintf(GetAllRowFormat, expectedName, expectedValue)))
}

func Test_GetAllHandler_NoValue_ReturnEmptyBody(t *testing.T) {
    // Given
    r := httptest.NewRequest(http.MethodGet, "/", nil)
    w := httptest.NewRecorder()

    service := mocks.NewServiceMock()
    handler := makeHandler(service)

    service.GetAllReturnValue = map[string]string{}

    // When
    handler.GetAll(w, r)

    // Then
    res := w.Result()
    defer res.Body.Close()

    body, err := io.ReadAll(res.Body)
    require.NoError(t, err, `failed to read response body`)

    require.Equal(t, res.StatusCode, http.StatusOK)
    require.Equal(t, body, []byte(``))
}

func Test_GetAllHandler_HasError_ThrowInternalError(t *testing.T) {
    // Given
    r := httptest.NewRequest(http.MethodGet, "/", nil)
    w := httptest.NewRecorder()

    service := mocks.NewServiceMock()
    handler := makeHandler(service)

    service.GetAllReturnError = errors.New(`test error`)

    // When
    handler.GetAll(w, r)

    // Then
    res := w.Result()
    defer res.Body.Close()

    body, err := io.ReadAll(res.Body)
    require.NoError(t, err, `failed to read response body`)

    require.Equal(t, body, []byte(``))
    require.Equal(t, res.StatusCode, http.StatusInternalServerError)
}

func Test_GetValueHandler_HasValue_Return(t *testing.T) {
    // Given
    expectedValue := `1400`

    r := httptest.NewRequest(http.MethodGet, "/value/gauge/test", nil)
    w := httptest.NewRecorder()

    r.SetPathValue(service.MetricTypePathKey, `gauge`)
    r.SetPathValue(service.MetricNamePathKey, `test`)

    service := mocks.NewServiceMock()
    handler := makeHandler(service)

    service.GetReturnValue = expectedValue

    // When
    handler.GetValue(w, r)

    // Then
    res := w.Result()
    defer res.Body.Close()

    body, err := io.ReadAll(res.Body)
    require.NoError(t, err, `failed to read response body`)

    require.Equal(t, res.StatusCode, http.StatusOK)
    require.Equal(t, body, []byte(expectedValue))
}

func Test_GetValueHandler_HasValue_ThrowNotFound(t *testing.T) {
    // Given
    r := httptest.NewRequest(http.MethodGet, "/value/gauge/test", nil)
    w := httptest.NewRecorder()

    r.SetPathValue(service.MetricTypePathKey, `gauge`)
    r.SetPathValue(service.MetricNamePathKey, `test`)

    service := mocks.NewServiceMock()
    handler := makeHandler(service)

    service.GetReturnValue = ``

    // When
    handler.GetValue(w, r)

    // Then
    res := w.Result()
    defer res.Body.Close()

    body, err := io.ReadAll(res.Body)
    require.NoError(t, err, `failed to read response body`)

    require.Equal(t, res.StatusCode, http.StatusNotFound)
    require.Equal(t, body, []byte(``))
}

func Test_GetValueHandler_HasError_ThrowInternalError(t *testing.T) {
    // Given
    r := httptest.NewRequest(http.MethodGet, "/value/gauge/test", nil)
    w := httptest.NewRecorder()

    r.SetPathValue(service.MetricTypePathKey, `gauge`)
    r.SetPathValue(service.MetricNamePathKey, `test`)

    service := mocks.NewServiceMock()
    handler := makeHandler(service)

    service.GetReturnError = errors.New(`test error`)

    // When
    handler.GetValue(w, r)

    // Then
    res := w.Result()
    defer res.Body.Close()

    body, err := io.ReadAll(res.Body)
    require.NoError(t, err, `failed to read response body`)

    require.Equal(t, res.StatusCode, http.StatusInternalServerError)
    require.Equal(t, body, []byte(``))
}
