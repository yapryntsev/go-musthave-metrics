package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMv "github.com/go-chi/chi/v5/middleware"
	"github.com/go-resty/resty/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yapryntsev/go-musthave-metrics/internal/handler"
	grpcSrv "github.com/yapryntsev/go-musthave-metrics/internal/handler/grpc"
	"github.com/yapryntsev/go-musthave-metrics/internal/middleware"
	grpcMdl "github.com/yapryntsev/go-musthave-metrics/internal/middleware/grpc"
	pb "github.com/yapryntsev/go-musthave-metrics/internal/proto"
	"github.com/yapryntsev/go-musthave-metrics/internal/repository"
	"github.com/yapryntsev/go-musthave-metrics/internal/service"
	"github.com/yapryntsev/go-musthave-metrics/internal/service/audit"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var log *zap.Logger

var buildVersion string = "N/A"
var buildDate string = "N/A"
var buildCommit string = "N/A"

func main() {
	_, _ = fmt.Fprintf(os.Stdout, "Build version: %s\n", buildVersion)
	_, _ = fmt.Fprintf(os.Stdout, "Build date: %s\n", buildDate)
	_, _ = fmt.Fprintf(os.Stdout, "Build commit: %s\n", buildCommit)

	setupLogger()

	err := parseFlags(os.Args[1:])
	if err != nil {
		log.Fatal("failed to launch app instance", zap.Error(err))
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)

	db := configureDB(ctx, log)
	svc := configureService(db, log)

	var httpServer *http.Server
	var grpcServer *grpc.Server

	if flagPreferGRPC {
		grpcServer = configureServerGRPC(svc, log)
	} else {
		httpServer = configureServerHTTP(flagAddr, svc, log)
	}

	serverError := make(chan error, 1)

	go func() {
		log.Debug("server is running")

		if flagPreferGRPC {
			listen, err := net.Listen("tcp", flagAddr)
			serverError <- err

			if err := grpcServer.Serve(listen); err != nil {
				serverError <- err
			}
		} else {
			if err := httpServer.ListenAndServe(); err != nil {
				serverError <- err
			}
		}
	}()

	select {
	case err := <-serverError:
		log.Debug("shutdown with server error", zap.Error(err))
	case <-ctx.Done():
		log.Debug("shutdown with os signal")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	defer func() {
		log.Debug("server terminated")
	}()

	defer func() {
		if flagPreferGRPC {
			grpcServer.Stop()
		} else {
			if err := httpServer.Shutdown(ctx); err != nil {
				log.Error("failed to gracefully shutdown server with error", zap.Error(err))
			}
		}
	}()

	defer func() {
		if db != nil {
			db.Close()
		}
	}()

	defer stop()
}

func configureDB(ctx context.Context, log *zap.Logger) *pgxpool.Pool {
	if flagDsn == "" {
		return nil
	}

	log.Debug(fmt.Sprintf("connect to DB, dsn: %s", flagDsn))

	config, err := pgxpool.ParseConfig(flagDsn)
	if err != nil {
		log.Fatal("failed to parse DSN", zap.Error(err))
		return nil
	}

	config.MaxConnIdleTime = 5 * time.Second
	config.MaxConnLifetime = 10 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatal("failed to create connection to db", zap.Error(err))
		return nil
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatal("failed to ping db", zap.Error(err))
	}

	return pool
}

func configureService(db *pgxpool.Pool, log *zap.Logger) *service.Service {
	metricRepo := repository.New(
		db,
		time.Duration(flagStoreInt),
		flagStorePath,
		flagRestore,
		log,
	)

	return service.New(metricRepo, db)
}

func configureServerGRPC(svc service.MetricService, logger *zap.Logger) *grpc.Server {
	server := grpc.NewServer(grpc.UnaryInterceptor(grpcMdl.SubnetInterceptor(flagTrustedSubnet, logger)))
	pb.RegisterMetricsServer(server, grpcSrv.NewServer(svc, logger))

	return server
}

func configureServerHTTP(addr string, svc service.MetricService, log *zap.Logger) *http.Server {
	log.Debug(fmt.Sprintf("server bootstrap, address: %s", addr))
	metricHandler := handler.New(svc, log)

	if flagAuditFile != "" {
		auditor, err := audit.NewLocalAuditor(flagAuditFile, log)
		if err != nil {
			log.Fatal("failed to initiate local auditor", zap.Error(err))
		}

		metricHandler.Audit(auditor)
	}

	if flagAuditURL != "" {
		httpClient := &http.Client{
			Timeout: 5 * time.Second,
		}
		restyClient := resty.NewWithClient(httpClient).
			SetRetryCount(3).
			SetRetryAfter(
				func(client *resty.Client, response *resty.Response) (time.Duration, error) {
					a := response.Request.Attempt - 1
					return time.Duration(1+2*a) * time.Second, nil
				},
			)
		auditor := audit.NewRemoteAuditor(flagAuditURL, restyClient, log)

		metricHandler.Audit(auditor)
	}

	privateKey, err := fetchPrivateKey()
	if err != nil {
		log.Fatal("failed to fetch private key", zap.Error(err))
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger(log))
	r.Use(middleware.Subnet(flagTrustedSubnet, log))
	r.Use(middleware.SignBody(flagSignKey, log))
	r.Use(middleware.Compress(log))
	r.Use(middleware.Decryptor(privateKey, log))

	r.Mount("/debug", chiMv.Profiler())

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
	r.Post("/updates", metricHandler.UpdateBatch)
	r.Post("/updates/", metricHandler.UpdateBatch)

	r.Get(getValueEndpoint, metricHandler.GetValue)
	r.Post(updateValueEndpoint, metricHandler.Update)

	log.Debug("handlers registered")

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

	log, err = zap.NewDevelopment()
	if err != nil {
		panic(fmt.Errorf("failed to initiate logger: %w", err))
	}
}

func fetchPrivateKey() (*rsa.PrivateKey, error) {
	if flagCryptoKey == "" {
		return nil, nil
	}

	file, err := os.ReadFile(flagCryptoKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	pemBlock, _ := pem.Decode(file)
	if pemBlock == nil {
		return nil, errors.New("failed to decode provided file")
	}

	key, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return key, nil
}
