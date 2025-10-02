package main

import (
    "flag"
    "fmt"
    "github.com/yapryntsev/go-musthave-metrics/internal/agent"
    "log"
    "os"
)

func main() {
    appLog := newLog("app")

    addr := new(string)
    reportInterval := new(uint)
    pollInterval := new(uint)

    parseFlags(addr, reportInterval, pollInterval)

    appLog.Println("agent bootstrap")
    metricAgent := agent.New(*addr, *reportInterval, *pollInterval, newLog(`agent`))

    appLog.Println("start metric gathering")
    if err := metricAgent.StartGathering(); err != nil {
        panic(err)
    }
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
