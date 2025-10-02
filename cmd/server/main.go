package main

import (
    "flag"
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
    addr := new(string)
    parseFlags(addr)

    if err := run(*addr); err != nil {
        panic(err)
    }
}

func parseFlags(addr *string) {
    flag.StringVar(addr, `a`, `localhost:8080`, `address and port to run server`)
    flag.Parse()
}

func run(addr string) error {
    appLog := newLog(`app`)

    appLog.Printf(`server bootstrap, address: %s`, addr)
    metricRepo := repository.NewInMemoryRepo()
    metricService := service.New(metricRepo, newLog("service"))
    metricHandler := handler.New(metricService, newLog("handler"))

    r := chi.NewRouter()
    r.Use(middleware.Timeout(5 * time.Second))

    getValueEndpoint := fmt.Sprintf(`/value/{%s}/{%s}`, service.MetricTypePathKey, service.MetricNamePathKey)
    updateValueEndpoint := fmt.Sprintf(
        `/update/{%s}/{%s}/{%s}`,
        service.MetricTypePathKey,
        service.MetricNamePathKey,
        service.MetricValuePathKey,
    )

    r.Get(`/`, metricHandler.GetAll)
    r.Get(getValueEndpoint, metricHandler.GetValue)
    r.Post(updateValueEndpoint, metricHandler.Update)

    appLog.Println(`handler registered`)

    return http.ListenAndServe(addr, r)
}

func newLog(prefix string) *log.Logger {
    return log.New(os.Stdout, fmt.Sprintf("%s: ", prefix), log.LstdFlags)
}
