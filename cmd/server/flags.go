package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	flagAddrDefault          = "localhost:8080"
	flagStoreIntDefault      = uint(300)
	flagStorePathDefault     = "./storage"
	flagRestoreDefault       = true
	flagDsnDefault           = ""
	flagSignKeyDefault       = ""
	flagAuditFileDefault     = ""
	flagAuditURLDefault      = ""
	flagCryptoKeyDefault     = ""
	flagConfigPathDefault    = ""
	flagTrustedSubnetDefault = ""
	flagPreferGRPCDefault    = false

	flagAddrKey          = "a"
	flagStoreIntKey      = "i"
	flagStorePathKey     = "f"
	flagRestoreKey       = "r"
	flagDsnKey           = "d"
	flagSignKeyKey       = "k"
	flagAuditFileKey     = "audit-file"
	flagAuditURLKey      = "audit-url"
	flagCryptoKeyKey     = "crypto-key"
	flagConfigPathKey    = "config"
	flagTrustedSubnetKey = "t"
	flagPreferGRPCKey    = "g"

	envAddrKey           = "ADDRESS"
	envStoreIntKey       = "STORE_INTERVAL"
	envStorePathKey      = "FILE_STORAGE_PATH"
	envRestoreKey        = "RESTORE"
	envDsnKey            = "DATABASE_DSN"
	envSignKeyKey        = "KEY"
	envAuditFileKey      = "AUDIT_FILE"
	envAuditURLKey       = "AUDIT_URL"
	envCryptoKeyKey      = "CRYPTO_KEY"
	envConfigPathKey     = "CONFIG"
	envTrustedSubnetKey  = "TRUSTED_SUBNET"
	envFlagPreferGRPCKey = "PREFER_GRPC"
)

var (
	flagAddr          string
	flagStoreInt      uint
	flagStorePath     string
	flagRestore       bool
	flagDsn           string
	flagSignKey       string
	flagAuditFile     string
	flagAuditURL      string
	flagCryptoKey     string
	flagTrustedSubnet string
	flagPreferGRPC    bool
)

type config struct {
	Address       string `json:"address"`
	Restore       bool   `json:"restore"`
	StoreInterval string `json:"store_interval"`
	StoreFile     string `json:"store_file"`
	DatabaseDsn   string `json:"database_dsn"`
	CryptoKey     string `json:"crypto_key"`
	TrustedSubnet string `json:"trusted_subnet"`
	PreferGRPC    bool   `json:"prefer_grpc"`
}

func parseFlags(args []string) error {
	var errs []error

	errs = append(errs, parseConfig(args))
	errs = append(errs, parseArgs(args))
	errs = append(errs, parseEnv())

	return errors.Join(errs...)
}

func parseConfig(args []string) error {
	var configPath string

	fs := flag.NewFlagSet("config-flag", flag.ContinueOnError)
	fs.StringVar(&configPath, flagConfigPathKey, flagConfigPathDefault, "config file path")

	err := fs.Parse(args)
	if err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	if envConfigPath, ok := os.LookupEnv(envConfigPathKey); ok {
		configPath = envConfigPath
	}

	var conf config
	if configPath != "" {
		file, err := os.Open(configPath)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}

		err = json.NewDecoder(file).Decode(&conf)
		if err != nil {
			return fmt.Errorf("failed to decode config file: %w", err)
		}
	}

	flagAddr = conf.Address
	flagRestore = conf.Restore
	flagStorePath = conf.StoreFile
	flagDsn = conf.DatabaseDsn
	flagCryptoKey = conf.CryptoKey
	flagTrustedSubnet = conf.TrustedSubnet
	flagPreferGRPC = conf.PreferGRPC

	val, err := time.ParseDuration(conf.StoreInterval)
	if err != nil {
		return fmt.Errorf("failed to decode report interval config value: %w", err)
	}

	flagStoreInt = uint(val.Seconds())
	return nil
}

func parseArgs(args []string) error {
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
	fs.StringVar(&flagTrustedSubnet, flagTrustedSubnetKey, flagTrustedSubnetDefault, "trusted subnet mask")
	fs.BoolVar(&flagPreferGRPC, flagPreferGRPCKey, flagPreferGRPCDefault, "should choose grpc over http")

	err := fs.Parse(args)
	if err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	return nil
}

func parseEnv() error {
	if envAddr, ok := os.LookupEnv(envAddrKey); ok {
		flagAddr = envAddr
	}

	if envStoreInt, ok := os.LookupEnv(envStoreIntKey); ok {
		d, err := strconv.Atoi(envStoreInt)
		if err != nil {
			return fmt.Errorf("failed to parse env value: %s", envStoreIntKey)
		}

		flagStoreInt = uint(d)
	}

	if envStorePath, ok := os.LookupEnv(envStorePathKey); ok {
		flagStorePath = envStorePath
	}

	if envRestore, ok := os.LookupEnv(envRestoreKey); ok {
		b, err := strconv.ParseBool(envRestore)
		if err != nil {
			return fmt.Errorf("failed to parse env value: %s", envRestoreKey)
		}

		flagRestore = b
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

	if envTrustedSubnet, ok := os.LookupEnv(envTrustedSubnetKey); ok {
		flagTrustedSubnet = envTrustedSubnet
	}

	if envPreferGRPC, ok := os.LookupEnv(envFlagPreferGRPCKey); ok {
		b, err := strconv.ParseBool(envPreferGRPC)
		if err != nil {
			return fmt.Errorf("failed to parse env value: %s", envFlagPreferGRPCKey)
		}

		flagPreferGRPC = b
	}

	return nil
}
