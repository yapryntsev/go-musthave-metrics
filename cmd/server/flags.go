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

func parseFlags(args []string, l *zap.Logger) {
    fs := flag.NewFlagSet("flags", flag.ExitOnError)

    fs.StringVar(&flagAddr, flagAddrKey, flagAddrDefault, `server endpoint`)
    fs.StringVar(&flagStorePath, flagStorePathKey, flagStorePathDefault, "file storage path")
    fs.BoolVar(&flagRestore, flagRestoreKey, flagRestoreDefault, "should restore storage state from file")
    fs.UintVar(&flagStoreInt, flagStoreIntKey, flagStoreIntDefault, "file write frequency")

    err := fs.Parse(args)
    if err != nil {
        l.Fatal("failed to parse flags", zap.Error(err))
    }

    if envAddr, ok := os.LookupEnv(envAddrKey); ok {
        flagAddr = envAddr
    }

    if envStoreInt, ok := os.LookupEnv(envStoreIntKey); ok {
        d, err := strconv.Atoi(envStoreInt)
        if err != nil {
            l.Error(fmt.Sprintf("failed to parse env value: %s", envStoreIntKey))
        } else {
            flagStoreInt = uint(d)
        }
    }

    if envStorePath, ok := os.LookupEnv(envStorePathKey); ok {
        flagStorePath = envStorePath
    }

    if envRestore, ok := os.LookupEnv(envRestoreKey); ok {
        b, err := strconv.ParseBool(envRestore)
        if err != nil {
            l.Error(fmt.Sprintf("failed to parse env value: %s", envRestoreKey))
        } else {
            flagRestore = b
        }
    }
}
