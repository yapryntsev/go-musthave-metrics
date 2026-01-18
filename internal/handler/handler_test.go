package handler

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    models "github.com/yapryntsev/go-musthave-metrics/internal/model"
    "github.com/yapryntsev/go-musthave-metrics/internal/service"
    "github.com/yapryntsev/go-musthave-metrics/internal/service/mocks"
    "go.uber.org/mock/gomock"
    "go.uber.org/zap/zaptest"
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
                ctrl := gomock.NewController(t)
                defer ctrl.Finish()

                r := httptest.NewRequest(m, "/gauge/test/100", nil)
                w := httptest.NewRecorder()

                // When
                handler := makeHandler(t, mocks.NewMockMetricService(ctrl))
                handler.Update(w, r)

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
                ctrl := gomock.NewController(t)
                defer ctrl.Finish()

                r := httptest.NewRequest(m, "/counter/test/100", nil)
                w := httptest.NewRecorder()

                // When
                handler := makeHandler(t, mocks.NewMockMetricService(ctrl))
                handler.Update(w, r)

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
                ctrl := gomock.NewController(t)
                defer ctrl.Finish()

                r := httptest.NewRequest(
                    http.MethodPost,
                    fmt.Sprintf(`/gauge/%s/%s`, c.name, c.value),
                    nil,
                )
                w := httptest.NewRecorder()

                r.SetPathValue(service.MetricNamePathKey, c.name)
                r.SetPathValue(service.MetricValuePathKey, c.value)

                // When
                handler := makeHandler(t, mocks.NewMockMetricService(ctrl))
                handler.Update(w, r)

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
                ctrl := gomock.NewController(t)
                defer ctrl.Finish()

                r := httptest.NewRequest(
                    http.MethodPost,
                    fmt.Sprintf(`/counter/%s/%s`, c.name, c.value),
                    nil,
                )
                w := httptest.NewRecorder()

                r.SetPathValue(service.MetricNamePathKey, c.name)
                r.SetPathValue(service.MetricValuePathKey, c.value)

                // When
                handler := makeHandler(t, mocks.NewMockMetricService(ctrl))
                handler.Update(w, r)

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

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    // When
    handler := makeHandler(t, mocks.NewMockMetricService(ctrl))
    handler.Update(w, r)

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

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    // When
    handler := makeHandler(t, mocks.NewMockMetricService(ctrl))
    handler.Update(w, r)

    // Then
    res := w.Result()
    defer res.Body.Close()

    assert.Equal(t, res.StatusCode, http.StatusBadRequest)
}

func Test_GetAllHandler_HasValue_Return(t *testing.T) {
    // Given
    expectedName := "test"
    expectedValue := "0"
    expectedReturn := []models.Metrics{
        {
            ID:    expectedName,
            MType: models.Counter,
            Delta: new(int64),
        },
    }

    r := httptest.NewRequest(http.MethodGet, "/", nil)
    w := httptest.NewRecorder()

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    svc := mocks.NewMockMetricService(ctrl)
    handler := makeHandler(t, svc)

    svc.EXPECT().
        GetAll(r.Context()).
        Return(expectedReturn, nil)

    // When
    handler.GetAll(w, r)

    // Then
    res := w.Result()
    defer res.Body.Close()

    body, err := io.ReadAll(res.Body)
    require.NoError(t, err, "failed to read response body")

    require.Equal(t, res.StatusCode, http.StatusOK)
    require.Equal(t, body, []byte(fmt.Sprintf(GetAllRowFormat, expectedName, expectedValue)))
}

func Test_GetAllHandler_NoValue_ReturnEmptyBody(t *testing.T) {
    // Given
    r := httptest.NewRequest(http.MethodGet, "/", nil)
    w := httptest.NewRecorder()

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    svc := mocks.NewMockMetricService(ctrl)
    handler := makeHandler(t, svc)

    svc.EXPECT().
        GetAll(r.Context()).
        Return([]models.Metrics{}, nil)

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

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    svc := mocks.NewMockMetricService(ctrl)
    handler := makeHandler(t, svc)

    svc.EXPECT().GetAll(r.Context()).Return([]models.Metrics{}, errors.New("test error"))

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
    expectedValue := float64(1400)

    r := httptest.NewRequest(http.MethodGet, "/value/gauge/test", nil)
    w := httptest.NewRecorder()

    r.SetPathValue(service.MetricTypePathKey, `gauge`)
    r.SetPathValue(service.MetricNamePathKey, `test`)

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    svc := mocks.NewMockMetricService(ctrl)
    handler := makeHandler(t, svc)

    svc.EXPECT().
        Get(r.Context(), gomock.Any()).
        DoAndReturn(
            func(ctx context.Context, m *models.Metrics) (bool, error) {
                m.Value = &expectedValue
                return true, nil
            },
        )

    // When
    handler.GetValue(w, r)

    // Then
    res := w.Result()
    defer res.Body.Close()

    body, err := io.ReadAll(res.Body)
    require.NoError(t, err, `failed to read response body`)

    require.Equal(t, res.StatusCode, http.StatusOK)
    require.Equal(t, body, []byte("1400"))
}

func Test_GetValueHandler_HasNotValue_ThrowNotFound(t *testing.T) {
    // Given
    r := httptest.NewRequest(http.MethodGet, "/value/gauge/test", nil)
    w := httptest.NewRecorder()

    r.SetPathValue(service.MetricTypePathKey, `gauge`)
    r.SetPathValue(service.MetricNamePathKey, `test`)

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    svc := mocks.NewMockMetricService(ctrl)
    handler := makeHandler(t, svc)

    svc.EXPECT().
        Get(r.Context(), gomock.Any()).
        Return(false, nil)

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

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    svc := mocks.NewMockMetricService(ctrl)
    handler := makeHandler(t, svc)

    svc.EXPECT().
        Get(r.Context(), gomock.Any()).
        Return(false, errors.New("test error"))

    // When
    handler.GetValue(w, r)

    // Then
    res := w.Result()
    defer res.Body.Close()

    body, err := io.ReadAll(res.Body)
    require.NoError(t, err, `failed to read response body`)

    require.Equal(t, res.StatusCode, http.StatusInternalServerError)
    require.Empty(t, body)
}

func Test_GetObjectHandler_HasError_ThrowInternalError(t *testing.T) {
    // Given
    requestBody, err := json.Marshal(
        models.Metrics{
            ID:    "test",
            MType: "gauge",
        },
    )

    if err != nil {
        t.Fatal(err)
    }

    r := httptest.NewRequest(http.MethodPost, "/value", bytes.NewBuffer(requestBody))
    w := httptest.NewRecorder()

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    svc := mocks.NewMockMetricService(ctrl)
    handler := makeHandler(t, svc)

    svc.EXPECT().
        Get(r.Context(), gomock.Any()).
        Return(false, errors.New("test error"))

    // When
    handler.GetObject(w, r)

    // Then
    res := w.Result()
    defer res.Body.Close()

    body, err := io.ReadAll(res.Body)
    require.NoError(t, err, `failed to read response body`)

    require.Equal(t, http.StatusInternalServerError, res.StatusCode)
    require.Empty(t, body)
}

func Test_GetObjectHandler_HasNotValue_ThrowNotFound(t *testing.T) {
    // Given
    requestBody, err := json.Marshal(
        models.Metrics{
            ID:    "test",
            MType: "gauge",
        },
    )

    if err != nil {
        t.Fatal(err)
    }

    r := httptest.NewRequest(http.MethodPost, "/value", bytes.NewBuffer(requestBody))
    w := httptest.NewRecorder()

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    svc := mocks.NewMockMetricService(ctrl)
    handler := makeHandler(t, svc)

    svc.EXPECT().
        Get(r.Context(), gomock.Any()).
        Return(false, nil)

    // When
    handler.GetObject(w, r)

    // Then
    res := w.Result()
    defer res.Body.Close()

    body, err := io.ReadAll(res.Body)
    require.NoError(t, err, `failed to read response body`)

    require.Equal(t, http.StatusNotFound, res.StatusCode)
    require.Equal(t, []byte(``), body)
}

func Test_GetObjectHandler_HasValue_Return(t *testing.T) {
    // Given
    expectedValue := float64(1400)
    expectedMetric := models.Metrics{
        ID:    "test",
        MType: "gauge",
        Value: &expectedValue,
    }
    expectedResp, err := json.Marshal(expectedMetric)
    if err != nil {
        t.Fatal(err)
    }

    requestBody, err := json.Marshal(
        models.Metrics{
            ID:    "test",
            MType: "gauge",
        },
    )
    if err != nil {
        t.Fatal(err)
    }

    r := httptest.NewRequest(http.MethodPost, "/value", bytes.NewBuffer(requestBody))
    w := httptest.NewRecorder()

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    svc := mocks.NewMockMetricService(ctrl)
    handler := makeHandler(t, svc)

    svc.EXPECT().
        Get(r.Context(), gomock.Any()).
        DoAndReturn(
            func(ctx context.Context, m *models.Metrics) (bool, error) {
                m.Value = &expectedValue
                return true, nil
            },
        )

    // When
    handler.GetObject(w, r)

    // Then
    res := w.Result()
    defer res.Body.Close()

    body, err := io.ReadAll(res.Body)
    require.NoError(t, err, `failed to read response body`)

    require.Equal(t, http.StatusOK, res.StatusCode)
    require.Equal(t, string(expectedResp)+"\n", string(body))
}

func Test_Ping_CallService(t *testing.T) {
    // Given
    r := httptest.NewRequest(http.MethodGet, "/ping", nil)
    w := httptest.NewRecorder()

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    svc := mocks.NewMockMetricService(ctrl)
    handler := makeHandler(t, svc)

    svc.EXPECT().
        Ping(r.Context()).
        Return(nil)

    // When
    handler.Ping(w, r)
}

func Test_Ping_HasError_Return500(t *testing.T) {
    // Given
    r := httptest.NewRequest(http.MethodGet, "/ping", nil)
    w := httptest.NewRecorder()

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    svc := mocks.NewMockMetricService(ctrl)
    handler := makeHandler(t, svc)

    svc.EXPECT().
        Ping(r.Context()).
        Return(errors.New("test error"))

    // When
    handler.Ping(w, r)

    // Then
    res := w.Result()
    defer res.Body.Close()

    assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func Test_Ping_NoError_Return200(t *testing.T) {
    // Given
    r := httptest.NewRequest(http.MethodGet, "/ping", nil)
    w := httptest.NewRecorder()

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    svc := mocks.NewMockMetricService(ctrl)
    handler := makeHandler(t, svc)

    svc.EXPECT().
        Ping(r.Context()).
        Times(1)

    // When
    handler.Ping(w, r)

    // Then
    res := w.Result()
    defer res.Body.Close()

    assert.Equal(t, http.StatusOK, res.StatusCode)
}

func Test_Updates_InvalidBody_Return400(t *testing.T) {
    // Given
    r := httptest.NewRequest(http.MethodPost, "/updates", nil)
    w := httptest.NewRecorder()

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    svc := mocks.NewMockMetricService(ctrl)
    handler := makeHandler(t, svc)

    svc.EXPECT().
        UpdateCounter(r.Context(), gomock.Any(), gomock.Any()).
        Times(0)

    // When
    handler.UpdateBatch(w, r)

    // Then
    res := w.Result()
    defer res.Body.Close()

    assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func Test_Updates_ValidBody_PassBatchToService(t *testing.T) {
    // Given
    expectedBatch := []*models.Metrics{
        {
            ID:    "test",
            MType: "gauge",
        },
        {
            ID:    "test-2",
            MType: "gauge",
        },
    }

    requestBody, err := json.Marshal(expectedBatch)
    if err != nil {
        t.Error(err)
    }

    r := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewBuffer(requestBody))
    w := httptest.NewRecorder()

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    svc := mocks.NewMockMetricService(ctrl)
    handler := makeHandler(t, svc)

    svc.EXPECT().
        UpdateCounter(r.Context(), gomock.Any(), gomock.Any()).
        Times(0)

    svc.EXPECT().
        UpdateBatch(r.Context(), expectedBatch)

    // When
    handler.UpdateBatch(w, r)

    // Then
    res := w.Result()
    defer res.Body.Close()

    assert.Equal(t, http.StatusOK, res.StatusCode)
}

func Test_Updates_ServiceReturnError_Return500(t *testing.T) {
    // Given
    expectedBatch := []models.Metrics{
        {
            ID:    "test",
            MType: "gauge",
        },
        {
            ID:    "test-2",
            MType: "gauge",
        },
    }

    requestBody, err := json.Marshal(expectedBatch)
    if err != nil {
        t.Error(err)
    }

    r := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewBuffer(requestBody))
    w := httptest.NewRecorder()

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    svc := mocks.NewMockMetricService(ctrl)
    handler := makeHandler(t, svc)

    svc.EXPECT().
        UpdateBatch(r.Context(), gomock.Any()).
        Return(errors.New("test error"))

    svc.EXPECT().
        UpdateCounter(r.Context(), gomock.Any(), gomock.Any()).
        Times(0)

    // When
    handler.UpdateBatch(w, r)

    // Then
    res := w.Result()
    defer res.Body.Close()

    assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func makeHandler(t *testing.T, service *mocks.MockMetricService) *MetricHandler {
    return &MetricHandler{
        log:     zaptest.NewLogger(t),
        service: service,
    }
}
