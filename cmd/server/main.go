package main

import (
    "context"
    "fmt"
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    log "github.com/sirupsen/logrus"
    "github.com/yapryntsev/go-musthave-metrics/internal/handler"
    "github.com/yapryntsev/go-musthave-metrics/internal/logger"
    "github.com/yapryntsev/go-musthave-metrics/internal/repository"
    "github.com/yapryntsev/go-musthave-metrics/internal/service"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func main() {
    appLogger := logger.New(`app`)

    err := parseFlags(os.Args[1:])
    if err != nil {
        appLogger.Fatal(err)
    }

    server := configureServer(flagAddr, appLogger)

    serverError := make(chan error, 1)
    stopSignal := make(chan os.Signal, 1)

    go func() {
        appLogger.Trace(`server is running`)
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

func configureServer(addr string, appLog *log.Entry) *http.Server {
    appLog.Printf(`server bootstrap, address: %s`, addr)
    metricRepo := repository.NewInMemoryRepo()
    metricService := service.New(metricRepo, logger.New("service"))
    metricHandler := handler.New(metricService, logger.New("handler"))

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

    handlersLogger := logger.New("handler")

    r.Get(`/`, logger.Middleware(metricHandler.GetAll, handlersLogger))

    r.Get("/value", logger.Middleware(metricHandler.GetObject, handlersLogger))
    r.Post("/value/", logger.Middleware(metricHandler.GetObject, handlersLogger))
    r.Post("/update", logger.Middleware(metricHandler.UpdateObject, handlersLogger))
    r.Post("/update/", logger.Middleware(metricHandler.UpdateObject, handlersLogger))

    r.Get(getValueEndpoint, logger.Middleware(metricHandler.GetValue, handlersLogger))
    r.Post(updateValueEndpoint, logger.Middleware(metricHandler.Update, handlersLogger))

    appLog.Println(`handlers registered`)

    return &http.Server{
        Addr:         addr,
        Handler:      r,
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 5 * time.Second,
        IdleTimeout:  30 * time.Second,
    }
}
