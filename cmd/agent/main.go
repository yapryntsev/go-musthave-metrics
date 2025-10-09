package main

import (
    "context"
    log "github.com/sirupsen/logrus"
    "github.com/yapryntsev/go-musthave-metrics/internal/agent"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    appLog := newLog("app")

    err := parseFlags(os.Args[1:], appLog)
    if err != nil {
        appLog.Fatal(err)
    }

    appAgent := configureAgent(appLog)
    ctx, cancel := context.WithCancel(context.Background())

    go func() {
        appLog.Println(`agent is running`)
        if err := appAgent.StartGathering(ctx); err != nil {
            appLog.Printf(`agent failed with error: %v`, err)
            os.Exit(1)
        }
    }()

    stopSignal := make(chan os.Signal, 1)
    signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)

    <-stopSignal

    appLog.Println(`shutdown the agent`)
    cancel()
}

func configureAgent(appLog *log.Entry) *agent.Agent {
    appLog.Println("agent bootstrap")
    return agent.New(flagAddr, flagReportInt, flagPollInt, newLog(`agent`))
}

func newLog(system string) *log.Entry {
    return log.WithField("system", system)
}
