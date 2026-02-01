package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/yapryntsev/go-musthave-metrics/internal/agent"
	"go.uber.org/zap"
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

	parseFlags(os.Args[1:], l)
	stopSignal := make(chan struct{}, 1)

	appAgent := configureAgent(l)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	go func() {
		l.Debug("agent is running")
		if err := appAgent.StartGathering(ctx); err != nil {
			l.Error("agent failed with error", zap.Error(err))
			stopSignal <- struct{}{}
		}
	}()

	go func() {
		<-ctx.Done()

		l.Debug("shutting down with os signal")
		stopSignal <- struct{}{}
	}()

	<-stopSignal

	l.Debug("shutdown the agent")
	cancel()
}

func configureAgent(l *zap.Logger) *agent.Agent {
	l.Debug("agent bootstrap")
	return agent.New(flagAddr, flagReportInt, flagPollInt, flagSignKey, l)
}
