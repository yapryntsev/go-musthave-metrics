package main

import (
    "context"
    "database/sql"
    "fmt"
    "github.com/go-chi/chi/v5"
    chiMiddleware "github.com/go-chi/chi/v5/middleware"
    "github.com/yapryntsev/go-musthave-metrics/internal/config/db"
    "github.com/yapryntsev/go-musthave-metrics/internal/handler"
    "github.com/yapryntsev/go-musthave-metrics/internal/middleware"
    "github.com/yapryntsev/go-musthave-metrics/internal/repository"
    "github.com/yapryntsev/go-musthave-metrics/internal/service"
    "go.uber.org/zap"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
)

var logger *zap.Logger

func main() {
    setupLogger()
    parseFlags(os.Args[1:], logger)

    server := configureServer(flagAddr, logger)

    serverError := make(chan error, 1)
    stopSignal := make(chan os.Signal, 1)

    go func() {
        logger.Debug("server is running")
        if err := server.ListenAndServe(); err != nil {
            serverError <- err
        }
    }()

    signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)

    select {
    case err := <-serverError:
        logger.Debug("shutdown with server error", zap.Error(err))
    case sig := <-stopSignal:
        logger.Debug(fmt.Sprintf("shutdown with os signal: %v", sig))
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        logger.Error("failed to gracefully shutdown server with error", zap.Error(err))
    }

    if err := db.CloseConnection(); err != nil {
        logger.Error("failed to close connection to db", zap.Error(err))
    }

    logger.Debug("server terminated")
}

func configureServer(addr string, l *zap.Logger) *http.Server {
    l.Debug(fmt.Sprintf("server bootstrap, address: %s", addr))

    var appDB *sql.DB
    var err error

    if flagDsn != "" {
        appDB, err = db.NewConnection(flagDsn)
    }
    if err != nil {
        l.Fatal("failed to create connection to db", zap.Error(err))
        return nil
    }

    metricRepo := repository.New(
        appDB,
        time.Duration(flagStoreInt),
        flagStorePath,
        flagRestore,
        l,
    )
    metricService := service.New(metricRepo)
    metricHandler := handler.New(metricService, appDB, l)

    r := chi.NewRouter()
    r.Use(middleware.Logger(l))
    r.Use(middleware.Compress(l))
    r.Use(chiMiddleware.Timeout(5 * time.Second))

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

    r.Get("/", metricHandler.GetAll)
    r.Get("/ping", metricHandler.Ping)

    r.Post("/value", metricHandler.GetObject)
    r.Post("/value/", metricHandler.GetObject)
    r.Post("/update", metricHandler.UpdateObject)
    r.Post("/update/", metricHandler.UpdateObject)

    r.Get(getValueEndpoint, metricHandler.GetValue)
    r.Post(updateValueEndpoint, metricHandler.Update)

    l.Debug("handlers registered")

    return &http.Server{
        Addr:         addr,
        Handler:      r,
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 5 * time.Second,
        IdleTimeout:  30 * time.Second,
    }
}

func setupLogger() {
    var err error

    logger, err = zap.NewDevelopment()
    if err != nil {
        panic(fmt.Errorf("failed to initiate logger: %w", err))
    }
}
