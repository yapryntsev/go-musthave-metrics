package main

import (
    "fmt"
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/yapryntsev/go-musthave-metrics/internal/handler"
    "github.com/yapryntsev/go-musthave-metrics/internal/repository"
    "github.com/yapryntsev/go-musthave-metrics/internal/service"
    "log"
    "net/http"
    "os"
    "time"
)

func main() {
    if err := run(8080); err != nil {
        panic(err)
    }
}

func run(port int) error {
    appLog := newLog(`app`)

    appLog.Printf(`server bootstrap, port :%d`, port)
    metricRepo := repository.NewInMemoryRepo()
    metricService := service.New(metricRepo, newLog("service"))
    metricHandler := handler.New(metricService, newLog("handler"))

    r := chi.NewRouter()
    r.Use(middleware.Timeout(5 * time.Second))

    r.Route(
        `/update`, func(r chi.Router) {
            metricTypes := [...]struct {
                name    string
                handler http.HandlerFunc
            }{
                {service.CounterMetricTypeName, metricHandler.UpdateCounter},
                {service.GaugeMetricTypeName, metricHandler.UpdateGauge},
            }

            for _, t := range metricTypes {
                endpoint := fmt.Sprintf(
                    `/%s/{%s}/{%s}`,
                    t.name,
                    service.MetricNamePathKey,
                    service.MetricValuePathKey,
                )

                r.Post(endpoint, t.handler)
            }
        },
    )

    getValueEndpoint := fmt.Sprintf(`/value/{%s}/{%s}`, service.MetricTypePathKey, service.MetricNamePathKey)
    r.Get(getValueEndpoint, metricHandler.GetValue)
    r.Get(`/`, metricHandler.GetAll)

    appLog.Println(`handler registered`)

    return http.ListenAndServe(fmt.Sprintf(`:%d`, port), r)
}

func newLog(prefix string) *log.Logger {
    return log.New(os.Stdout, fmt.Sprintf("%s: ", prefix), log.LstdFlags)
}
