package main

import (
    "flag"
    "os"
)

const (
    flagAddrDefault = "localhost:8080"
    flagAddrKey     = "a"
    envAddrKey      = "ADDRESS"
)

var flagAddr string

func parseFlags(args []string) error {
    fs := flag.NewFlagSet("flags", flag.ExitOnError)
    fs.StringVar(&flagAddr, flagAddrKey, flagAddrDefault, `server endpoint`)

    err := fs.Parse(args)
    if err != nil {
        return err
    }

    if envAddr := os.Getenv(envAddrKey); envAddr != "" {
        flagAddr = envAddr
    }

    return nil
}
