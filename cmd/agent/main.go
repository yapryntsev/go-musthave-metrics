package main

import (
    "context"
    "fmt"
    "github.com/yapryntsev/go-musthave-metrics/internal/agent"
    "go.uber.org/zap"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    l, err := zap.NewDevelopment()
    if err != nil {
        panic(fmt.Errorf("failed to initiate logger: %w", err))
    }

    parseFlags(os.Args[1:], l)
    stopSignal := make(chan struct{}, 1)

    appAgent := configureAgent(l)
    ctx, cancel := context.WithCancel(context.Background())

    go func() {
        l.Debug("agent is running")
        if err := appAgent.StartGathering(ctx); err != nil {
            l.Error("agent failed with error", zap.Error(err))
            stopSignal <- struct{}{}
        }
    }()

    go func() {
        ch := make(chan os.Signal, 1)

        signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
        sig := <-ch

        l.Debug("shutting down with os signal", zap.String("signal", sig.String()))
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
