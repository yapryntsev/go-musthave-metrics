package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
)

const (
	flagAddrDefault       = "localhost:8080"
	flagStoreIntDefault   = uint(300)
	flagStorePathDefault  = "./storage"
	flagRestoreDefault    = true
	flagDsnDefault        = ""
	flagSignKeyDefault    = ""
	flagAuditFileDefault  = ""
	flagAuditURLDefault   = ""
	flagCryptoKeyDefault  = ""
	flagConfigPathDefault = ""

	flagAddrKey       = "a"
	flagStoreIntKey   = "i"
	flagStorePathKey  = "f"
	flagRestoreKey    = "r"
	flagDsnKey        = "d"
	flagSignKeyKey    = "k"
	flagAuditFileKey  = "audit-file"
	flagAuditURLKey   = "audit-url"
	flagCryptoKeyKey  = "crypto-key"
	flagConfigPathKey = "c"

	envAddrKey       = "ADDRESS"
	envStoreIntKey   = "STORE_INTERVAL"
	envStorePathKey  = "FILE_STORAGE_PATH"
	envRestoreKey    = "RESTORE"
	envDsnKey        = "DATABASE_DSN"
	envSignKeyKey    = "KEY"
	envAuditFileKey  = "AUDIT_FILE"
	envAuditURLKey   = "AUDIT_URL"
	envCryptoKeyKey  = "CRYPTO_KEY"
	envConfigPathKey = "CONFIG"
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
	flagCryptoKey string
)

type config struct {
	Address       string `json:"address"`
	Restore       bool   `json:"restore"`
	StoreInterval string `json:"store_interval"`
	StoreFile     string `json:"store_file"`
	DatabaseDsn   string `json:"database_dsn"`
	CryptoKey     string `json:"crypto_key"`
}

func parseFlags(args []string, log *zap.Logger) {
	parseConfig(log)
	parseArgs(args, log)
	parseEnv(log)
}

func parseConfig(log *zap.Logger) {
	var configPath string

	flag.StringVar(&configPath, flagConfigPathKey, flagConfigPathDefault, "config file path")

	var conf config
	if configPath == "" {
		file, err := os.Open(configPath)
		if err != nil {
			log.Fatal("failed to read config file", zap.Error(err))
		}

		err = json.NewDecoder(file).Decode(&conf)
		if err != nil {
			log.Fatal("failed to decode config file", zap.Error(err))
		}
	}

	if envConfigPath, ok := os.LookupEnv(envConfigPathKey); ok {
		configPath = envConfigPath
	}

	flagAddr = conf.Address
	flagRestore = conf.Restore
	flagStorePath = conf.StoreFile
	flagDsn = conf.DatabaseDsn
	flagCryptoKey = conf.CryptoKey

	val, err := time.ParseDuration(conf.StoreInterval)
	if err != nil {
		log.Fatal("failed to decode report interval config value", zap.Error(err))
	}
	flagStoreInt = uint(val.Seconds())
}

func parseArgs(args []string, log *zap.Logger) {
	fs := flag.NewFlagSet("flags", flag.ExitOnError)

	fs.StringVar(&flagAddr, flagAddrKey, flagAddrDefault, `server endpoint`)
	fs.StringVar(&flagStorePath, flagStorePathKey, flagStorePathDefault, "file storage path")
	fs.BoolVar(&flagRestore, flagRestoreKey, flagRestoreDefault, "should restore storage state from file")
	fs.UintVar(&flagStoreInt, flagStoreIntKey, flagStoreIntDefault, "file write frequency")
	fs.StringVar(&flagDsn, flagDsnKey, flagDsnDefault, "data source name")
	fs.StringVar(&flagSignKey, flagSignKeyKey, flagSignKeyDefault, "key used to sign request body")
	fs.StringVar(&flagAuditFile, flagAuditFileKey, flagAuditFileDefault, "audit file")
	fs.StringVar(&flagAuditURL, flagAuditURLKey, flagAuditURLDefault, "audit remote service url")
	fs.StringVar(&flagCryptoKey, flagCryptoKeyKey, flagCryptoKeyDefault, "crypto key file path")

	err := fs.Parse(args)
	if err != nil {
		log.Fatal("failed to parse flags", zap.Error(err))
	}
}

func parseEnv(log *zap.Logger) {
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

	if envCryptoKey, ok := os.LookupEnv(envCryptoKeyKey); ok {
		flagCryptoKey = envCryptoKey
	}
}
