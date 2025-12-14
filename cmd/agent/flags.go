package main

import (
    "flag"
    "fmt"
    "go.uber.org/zap"
    "os"
    "strconv"
)

const (
    flagAddrDefault      = "localhost:8080"
    flagReportIntDefault = uint(10)
    flagPollIntDefault   = uint(2)
    flagSignKeyDefault   = ""

    flagAddrKey      = "a"
    flagReportIntKey = "r"
    flagPollIntKey   = "p"
    flagSignKeyKey   = "k"

    envAddrKey      = "ADDRESS"
    envReportIntKey = "REPORT_INTERVAL"
    envPollIntKey   = "POLL_INTERVAL"
    envSignKeyKey   = "KEY"
)

var (
    flagAddr      string
    flagReportInt uint
    flagPollInt   uint
    flagSignKey   string
)

func parseFlags(args []string, l *zap.Logger) {
    fs := flag.NewFlagSet("flags", flag.ExitOnError)

    fs.StringVar(&flagAddr, flagAddrKey, flagAddrDefault, `server endpoint`)
    fs.UintVar(&flagPollInt, flagPollIntKey, flagPollIntDefault, `frequency of gathering metrics in sec`)
    fs.StringVar(&flagSignKey, flagSignKeyKey, flagSignKeyDefault, "key used to sign request body")
    fs.UintVar(
        &flagReportInt,
        flagReportIntKey,
        flagReportIntDefault,
        `frequency of sending metrics to the server in sec`,
    )

    err := fs.Parse(args)
    if err != nil {
        l.Fatal("failed to parse flags", zap.Error(err))
    }

    if envAddr, ok := os.LookupEnv(envAddrKey); ok {
        flagAddr = envAddr
    }

    if envReportInt, ok := os.LookupEnv(envReportIntKey); ok {
        f, err := strconv.Atoi(envReportInt)
        if err != nil {
            l.Error(fmt.Sprintf("failed to parse env value: %s", envReportInt))
        } else {
            flagReportInt = uint(f)
        }
    }

    if envPollInt, ok := os.LookupEnv(envPollIntKey); ok {
        f, err := strconv.Atoi(envPollInt)
        if err != nil {
            l.Error(fmt.Sprintf("failed to parse env value: %s", envPollIntKey))
        } else {
            flagPollInt = uint(f)
        }
    }

    if signKey, ok := os.LookupEnv(envSignKeyKey); ok {
        flagSignKey = signKey
    }
}
