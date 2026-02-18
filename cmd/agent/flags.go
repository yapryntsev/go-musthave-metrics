package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
)

const (
	flagAddrDefault       = "localhost:8080"
	flagReportIntDefault  = uint(10)
	flagPollIntDefault    = uint(2)
	flagSignKeyDefault    = ""
	flagCryptoKeyDefault  = ""
	flagConfigPathDefault = ""

	flagAddrKey       = "a"
	flagReportIntKey  = "r"
	flagPollIntKey    = "p"
	flagSignKeyKey    = "k"
	flagCryptoKeyKey  = "crypto-key"
	flagConfigPathKey = "c"

	envAddrKey       = "ADDRESS"
	envReportIntKey  = "REPORT_INTERVAL"
	envPollIntKey    = "POLL_INTERVAL"
	envSignKeyKey    = "KEY"
	envCryptoKeyKey  = "CRYPTO_KEY"
	envConfigPathKey = "CONFIG"
)

var (
	flagAddr      string
	flagReportInt uint
	flagPollInt   uint
	flagSignKey   string
	flagCryptoKey string
)

type config struct {
	Address        string `json:"address"`
	ReportInterval string `json:"report_interval"`
	PollInterval   string `json:"poll_interval"`
	CryptoKey      string `json:"crypto_key"`
}

func parseFlags(args []string, l *zap.Logger) {
	parseConfig(l)
	parseArgs(args, l)
	parseEnv(l)
}

func parseConfig(l *zap.Logger) {
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
	flagCryptoKey = conf.CryptoKey

	val, err := time.ParseDuration(conf.ReportInterval)
	if err != nil {
		l.Fatal("failed to decode report interval config value", zap.Error(err))
	}
	flagReportInt = uint(val.Seconds())

	val, err = time.ParseDuration(conf.PollInterval)
	if err != nil {
		l.Fatal("failed to decode report interval config value", zap.Error(err))
	}
	flagPollInt = uint(val.Seconds())
}

func parseArgs(args []string, l *zap.Logger) {
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

	err := fs.Parse(args)
	if err != nil {
		l.Fatal("failed to parse flags", zap.Error(err))
	}
}

func parseEnv(l *zap.Logger) {
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

	if envCryptoKey, ok := os.LookupEnv(envCryptoKeyKey); ok {
		flagCryptoKey = envCryptoKey
	}
}
