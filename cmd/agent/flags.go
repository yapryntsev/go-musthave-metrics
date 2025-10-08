package main

import (
    "flag"
    "log"
    "os"
    "strconv"
)

const (
    flagAddrDefault      = "localhost:8080"
    flagReportIntDefault = uint(10)
    flagPollIntDefault   = uint(2)

    flagAddrKey      = "a"
    flagReportIntKey = "r"
    flagPollIntKey   = "p"

    envAddrKey      = "ADDRESS"
    envReportIntKey = "REPORT_INTERVAL"
    envPollIntKey   = "POLL_INTERVAL"
)

var (
    flagAddr      string
    flagReportInt uint
    flagPollInt   uint
)

func parseFlags(args []string, l *log.Logger) error {
    fs := flag.NewFlagSet("flags", flag.ExitOnError)

    fs.StringVar(&flagAddr, flagAddrKey, flagAddrDefault, `server endpoint`)
    fs.UintVar(&flagPollInt, flagPollIntKey, flagPollIntDefault, `frequency of gathering metrics in sec`)
    fs.UintVar(
        &flagReportInt,
        flagReportIntKey,
        flagReportIntDefault,
        `frequency of sending metrics to the server in sec`,
    )

    err := fs.Parse(args)
    if err != nil {
        return err
    }

    if envAddr := os.Getenv(envAddrKey); envAddr != "" {
        flagAddr = envAddr
    }

    if envReportInt := os.Getenv(envReportIntKey); envReportInt != "" {
        f, err := strconv.Atoi(envReportInt)
        if err != nil {
            l.Printf("failed to parse env value: %s", envReportIntKey)
        } else {
            flagReportInt = uint(f)
        }
    }

    if envPollInt := os.Getenv(envPollIntKey); envPollInt != "" {
        f, err := strconv.Atoi(envPollInt)
        if err != nil {
            l.Printf("failed to parse env value: %s", envPollIntKey)
        } else {
            flagPollInt = uint(f)
        }
    }

    return nil
}
