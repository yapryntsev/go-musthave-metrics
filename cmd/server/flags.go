package main

import (
    "flag"
    log "github.com/sirupsen/logrus"
    "os"
    "strconv"
)

const (
    flagAddrDefault      = "localhost:8080"
    flagStoreIntDefault  = uint(300)
    flagStorePathDefault = "./storage"
    flagRestoreDefault   = true

    flagAddrKey      = "a"
    flagStoreIntKey  = "i"
    flagStorePathKey = "f"
    flagRestoreKey   = "r"

    envAddrKey      = "ADDRESS"
    envStoreIntKey  = "STORE_INTERVAL"
    envStorePathKey = "FILE_STORAGE_PATH"
    envRestoreKey   = "RESTORE"
)

var (
    flagAddr      string
    flagStoreInt  uint
    flagStorePath string
    flagRestore   bool
)

func parseFlags(args []string, l *log.Entry) error {
    fs := flag.NewFlagSet("flags", flag.ExitOnError)
    fs.StringVar(&flagAddr, flagAddrKey, flagAddrDefault, `server endpoint`)
    fs.StringVar(&flagStorePath, flagStorePathKey, flagStorePathDefault, "file storage path")
    fs.BoolVar(&flagRestore, flagRestoreKey, flagRestoreDefault, "should restore storage state from file")
    fs.UintVar(&flagStoreInt, flagStoreIntKey, flagStoreIntDefault, "file write frequency")

    err := fs.Parse(args)
    if err != nil {
        return err
    }

    if envAddr := os.Getenv(envAddrKey); envAddr != "" {
        flagAddr = envAddr
    }

    if envStoreInt := os.Getenv(envStoreIntKey); envStoreInt != "" {
        d, err := strconv.Atoi(envStoreInt)
        if err != nil {
            l.Printf("failed to parse env value: %s", envStoreIntKey)
        } else {
            flagStoreInt = uint(d)
        }
    }

    if envStorePath := os.Getenv(envStorePathKey); envStorePath != "" {
        flagStorePath = envStorePath
    }

    if envRestore := os.Getenv(envRestoreKey); envRestore != "" {
        b, err := strconv.ParseBool(envRestore)
        if err != nil {
            l.Printf("failed to parse env value: %s", envRestoreKey)
        } else {
            flagRestore = b
        }
    }

    return nil
}
