package main

import (
    "fmt"
    "github.com/yapryntsev/go-musthave-metrics/internal/handler"
    "github.com/yapryntsev/go-musthave-metrics/internal/repository"
    "github.com/yapryntsev/go-musthave-metrics/internal/service"
    "log"
    "net/http"
    "os"
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

    mux := http.NewServeMux()
    mux.HandleFunc(
        fmt.Sprintf(
            `/update/{%s}/{%s}/{%s}`,
            service.MetricTypePathKey,
            service.MetricNamePathKey,
            service.MetricValuePathKey,
        ),
        metricHandler.Update,
    )
    appLog.Println(`handler registered`)

    return http.ListenAndServe(fmt.Sprintf(`:%d`, port), mux)
}

func newLog(prefix string) *log.Logger {
    return log.New(os.Stdout, fmt.Sprintf("%s: ", prefix), log.LstdFlags)
}
