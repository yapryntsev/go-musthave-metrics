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
	flagReportIntDefault = uint(10)
	flagPollIntDefault   = uint(2)
	flagSignKeyDefault   = ""
	flagCryptoKeyDefault = ""

	flagAddrKey      = "a"
	flagReportIntKey = "r"
	flagPollIntKey   = "p"
	flagSignKeyKey   = "k"
	flagCryptoKeyKey = "crypto-key"

	envAddrKey      = "ADDRESS"
	envReportIntKey = "REPORT_INTERVAL"
	envPollIntKey   = "POLL_INTERVAL"
	envSignKeyKey   = "KEY"
	envCryptoKeyKey = "CRYPTO_KEY"
)

var (
	flagAddr      string
	flagReportInt uint
	flagPollInt   uint
	flagSignKey   string
	flagCryptoKey string
)

func parseFlags(args []string, l *zap.Logger) {
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
