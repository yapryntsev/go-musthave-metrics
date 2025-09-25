package main

import (
    "fmt"
    "github.com/yapryntsev/go-musthave-metrics/internal/agent"
    "log"
    "os"
)

func main() {
    appLog := newLog("app")

    appLog.Println("agent bootstrap")
    metricAgent := agent.New("localhost", 8080, newLog("agent"))

    appLog.Println("start metric gathering")
    if err := metricAgent.StartGathering(); err != nil {
        panic(err)
    }
}

func newLog(prefix string) *log.Logger {
    return log.New(os.Stdout, fmt.Sprintf("%s: ", prefix), log.LstdFlags)
}
