package main

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yapryntsev/go-musthave-metrics/internal/agent"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

var buildVersion string = "N/A"
var buildDate string = "N/A"
var buildCommit string = "N/A"

func main() {
	_, _ = fmt.Fprintf(os.Stdout, "Build version: %s\n", buildVersion)
	_, _ = fmt.Fprintf(os.Stdout, "Build date: %s\n", buildDate)
	_, _ = fmt.Fprintf(os.Stdout, "Build commit: %s\n", buildCommit)

	l, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to initiate logger: %w", err))
	}

	err = parseFlags(os.Args[1:])
	if err != nil {
		l.Fatal("failed to launch app instance", zap.Error(err))
	}

	appAgent := configureAgent(l)
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)

	group, ctx := errgroup.WithContext(ctx)
	group.Go(
		func() error {
			l.Debug("agent is running")
			return appAgent.StartGathering(ctx)
		},
	)

	if err := group.Wait(); err != nil {
		l.Error("agent failed with error", zap.Error(err))
	}

	l.Debug("shutdown the agent")
	cancel()
}

func configureAgent(l *zap.Logger) *agent.Agent {
	l.Debug("agent bootstrap")

	cert, err := fetchCert()
	if err != nil {
		l.Fatal("failed to fetch cert", zap.Error(err))
	}

	return agent.New(flagAddr, flagReportInt, flagPollInt, flagSignKey, cert, l)
}

func fetchCert() (*x509.Certificate, error) {
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

	cert, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}
