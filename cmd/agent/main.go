package main

import (
    "context"
    "flag"
    "fmt"
    "github.com/yapryntsev/go-musthave-metrics/internal/agent"
    "log"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    appLog := newLog("app")
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

func configureAgent(appLog *log.Logger) *agent.Agent {
    addr := new(string)
    reportInterval := new(uint)
    pollInterval := new(uint)

    parseFlags(addr, reportInterval, pollInterval)

    appLog.Println("agent bootstrap")
    return agent.New(*addr, *reportInterval, *pollInterval, newLog(`agent`))
}

func parseFlags(addr *string, reportInterval *uint, pollInterval *uint) {
    flag.StringVar(addr, `a`, `localhost:8080`, `server endpoint`)
    flag.UintVar(reportInterval, `r`, 10, `frequency of sending metrics to the server in sec`)
    flag.UintVar(pollInterval, `p`, 2, `frequency of gathering metrics in sec`)
    flag.Parse()
}

func newLog(prefix string) *log.Logger {
    return log.New(os.Stdout, fmt.Sprintf("%s: ", prefix), log.LstdFlags)
}
