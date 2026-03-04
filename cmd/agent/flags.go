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
	flagAddrDefault       = "localhost:8080"
	flagReportIntDefault  = uint(10)
	flagPollIntDefault    = uint(2)
	flagSignKeyDefault    = ""
	flagCryptoKeyDefault  = ""
	flagConfigPathDefault = ""
	flagPreferGRPCDefault = false

	flagAddrKey       = "a"
	flagReportIntKey  = "r"
	flagPollIntKey    = "p"
	flagSignKeyKey    = "k"
	flagCryptoKeyKey  = "crypto-key"
	flagConfigPathKey = "config"
	flagPreferGRPCKey = "g"

	envAddrKey           = "ADDRESS"
	envReportIntKey      = "REPORT_INTERVAL"
	envPollIntKey        = "POLL_INTERVAL"
	envSignKeyKey        = "KEY"
	envCryptoKeyKey      = "CRYPTO_KEY"
	envConfigPathKey     = "CONFIG"
	envFlagPreferGRPCKey = "PREFER_GRPC"
)

var (
	flagAddr       string
	flagReportInt  uint
	flagPollInt    uint
	flagSignKey    string
	flagCryptoKey  string
	flagPreferGRPC bool
)

type config struct {
	Address        string `json:"address"`
	ReportInterval string `json:"report_interval"`
	PollInterval   string `json:"poll_interval"`
	CryptoKey      string `json:"crypto_key"`
	PreferGRPC     bool   `json:"prefer_grpc"`
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
	flagCryptoKey = conf.CryptoKey
	flagPreferGRPC = conf.PreferGRPC

	val, err := time.ParseDuration(conf.ReportInterval)
	if err != nil {
		return fmt.Errorf("failed to decode report interval config value: %w", err)
	}
	flagReportInt = uint(val.Seconds())

	val, err = time.ParseDuration(conf.PollInterval)
	if err != nil {
		return fmt.Errorf("failed to decode report interval config value: %w", err)
	}
	flagPollInt = uint(val.Seconds())

	return nil
}

func parseArgs(args []string) error {
	fs := flag.NewFlagSet("flags", flag.ExitOnError)

	fs.StringVar(&flagAddr, flagAddrKey, flagAddrDefault, `server endpoint`)
	fs.UintVar(&flagPollInt, flagPollIntKey, flagPollIntDefault, `frequency of gathering metrics in sec`)
	fs.StringVar(&flagSignKey, flagSignKeyKey, flagSignKeyDefault, "key used to sign request body")
	fs.StringVar(&flagCryptoKey, flagCryptoKeyKey, flagCryptoKeyDefault, "crypto key file path")
	fs.UintVar(
		&flagReportInt,
		flagReportIntKey,
		flagReportIntDefault,
		`frequency of sending metrics to the server in sec`,
	)
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

	if envReportInt, ok := os.LookupEnv(envReportIntKey); ok {
		f, err := strconv.Atoi(envReportInt)
		if err != nil {
			return fmt.Errorf("failed to parse env value: %s", envReportInt)
		}

		flagReportInt = uint(f)
	}

	if envPollInt, ok := os.LookupEnv(envPollIntKey); ok {
		f, err := strconv.Atoi(envPollInt)
		if err != nil {
			return fmt.Errorf("failed to parse env value: %s", envPollIntKey)
		}

		flagPollInt = uint(f)
	}

	if signKey, ok := os.LookupEnv(envSignKeyKey); ok {
		flagSignKey = signKey
	}

	if envCryptoKey, ok := os.LookupEnv(envCryptoKeyKey); ok {
		flagCryptoKey = envCryptoKey
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
