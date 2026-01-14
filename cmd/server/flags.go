package main

import (
    "flag"
    "fmt"
    "os"
    "strconv"

    "go.uber.org/zap"
)

const (
    flagAddrDefault      = "localhost:8080"
    flagStoreIntDefault  = uint(300)
    flagStorePathDefault = "./storage"
    flagRestoreDefault   = true
    flagDsnDefault       = ""
    flagSignKeyDefault   = ""
    flagAuditFileDefault = ""
    flagAuditURLDefault  = ""

    flagAddrKey      = "a"
    flagStoreIntKey  = "i"
    flagStorePathKey = "f"
    flagRestoreKey   = "r"
    flagDsnKey       = "d"
    flagSignKeyKey   = "k"
    flagAuditFileKey = "audit-file"
    flagAuditURLKey  = "audit-url"

    envAddrKey      = "ADDRESS"
    envStoreIntKey  = "STORE_INTERVAL"
    envStorePathKey = "FILE_STORAGE_PATH"
    envRestoreKey   = "RESTORE"
    envDsnKey       = "DATABASE_DSN"
    envSignKeyKey   = "KEY"
    envAuditFileKey = "AUDIT_FILE"
    envAuditURLKey  = "AUDIT_URL"
)

var (
    flagAddr      string
    flagStoreInt  uint
    flagStorePath string
    flagRestore   bool
    flagDsn       string
    flagSignKey   string
    flagAuditFile string
    flagAuditURL  string
)

func parseFlags(args []string, log *zap.Logger) {
    fs := flag.NewFlagSet("flags", flag.ExitOnError)

    fs.StringVar(&flagAddr, flagAddrKey, flagAddrDefault, `server endpoint`)
    fs.StringVar(&flagStorePath, flagStorePathKey, flagStorePathDefault, "file storage path")
    fs.BoolVar(&flagRestore, flagRestoreKey, flagRestoreDefault, "should restore storage state from file")
    fs.UintVar(&flagStoreInt, flagStoreIntKey, flagStoreIntDefault, "file write frequency")
    fs.StringVar(&flagDsn, flagDsnKey, flagDsnDefault, "data source name")
    fs.StringVar(&flagSignKey, flagSignKeyKey, flagSignKeyDefault, "key used to sign request body")
    fs.StringVar(&flagAuditFile, flagAuditFileKey, flagAuditFileDefault, "audit file")
    fs.StringVar(&flagAuditURL, flagAuditURLKey, flagAuditURLDefault, "audit remote service url")

    err := fs.Parse(args)
    if err != nil {
        log.Fatal("failed to parse flags", zap.Error(err))
    }

    if envAddr, ok := os.LookupEnv(envAddrKey); ok {
        flagAddr = envAddr
    }

    if envStoreInt, ok := os.LookupEnv(envStoreIntKey); ok {
        d, err := strconv.Atoi(envStoreInt)
        if err != nil {
            log.Error(fmt.Sprintf("failed to parse env value: %s", envStoreIntKey))
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
            log.Error(fmt.Sprintf("failed to parse env value: %s", envRestoreKey))
        } else {
            flagRestore = b
        }
    }

    if envDsn, ok := os.LookupEnv(envDsnKey); ok {
        flagDsn = envDsn
    }

    if signKey, ok := os.LookupEnv(envSignKeyKey); ok {
        flagSignKey = signKey
    }

    if auditFile, ok := os.LookupEnv(envAuditFileKey); ok {
        flagAuditFile = auditFile
    }

    if auditURL, ok := os.LookupEnv(envAuditURLKey); ok {
        flagAuditURL = auditURL
    }
}
