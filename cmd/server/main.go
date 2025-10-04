package main

import (
    "context"
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
    "os/signal"
    "syscall"
    "time"
)

func main() {
    appLog := newLog(`app`)

    addr := new(string)
    parseFlags(addr)

    server := configureServer(*addr, appLog)

    serverError := make(chan error, 1)
    stopSignal := make(chan os.Signal, 1)

    go func() {
        appLog.Println(`server is running`)
        if err := server.ListenAndServe(); err != nil {
            serverError <- err
        }
    }()

    signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)

    select {
    case err := <-serverError:
        log.Printf(`shutdown with server error: %v`, err)
    case sig := <-stopSignal:
        log.Printf(`shutdown with os signal: %v`, sig)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        log.Printf(`failed to gracefully shutdown server with error: %v`, err)
    }

    log.Println(`server terminated`)
}

func parseFlags(addr *string) {
    flag.StringVar(addr, `a`, `localhost:8080`, `address and port to run server`)
    flag.Parse()
}

func configureServer(addr string, appLog *log.Logger) *http.Server {
    appLog.Printf(`server bootstrap, address: %s`, addr)
    metricRepo := repository.NewInMemoryRepo()
    metricService := service.New(metricRepo, newLog("service"))
    metricHandler := handler.New(metricService, newLog("handler"))

    r := chi.NewRouter()
    r.Use(middleware.Timeout(5 * time.Second))

    getValueEndpoint := fmt.Sprintf(
        `/value/{%s}/{%s}`,
        service.MetricTypePathKey,
        service.MetricNamePathKey,
    )
    updateValueEndpoint := fmt.Sprintf(
        `/update/{%s}/{%s}/{%s}`,
        service.MetricTypePathKey,
        service.MetricNamePathKey,
        service.MetricValuePathKey,
    )

    r.Get(`/`, metricHandler.GetAll)
    r.Get(getValueEndpoint, metricHandler.GetValue)
    r.Post(updateValueEndpoint, metricHandler.Update)

    appLog.Println(`handlers registered`)

    return &http.Server{
        Addr:         addr,
        Handler:      r,
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 5 * time.Second,
        IdleTimeout:  30 * time.Second,
    }
}

func newLog(prefix string) *log.Logger {
    return log.New(os.Stdout, fmt.Sprintf("%s: ", prefix), log.LstdFlags)
}
