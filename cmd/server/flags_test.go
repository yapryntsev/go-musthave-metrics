package main

import (
    "fmt"
    log "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/require"
    "testing"
)

func Test_ParseEnv(t *testing.T) {
    tests := []struct {
        envName  string
        envValue string
        expected interface{}
        actual   func() interface{}
    }{
        {envAddrKey, ":8080", ":8080", func() interface{} { return flagAddr }},
        {envStoreIntKey, "120", uint(120), func() interface{} { return flagStoreInt }},
        {envStorePathKey, "/test", "/test", func() interface{} { return flagStorePath }},
        {envRestoreKey, "false", false, func() interface{} { return flagRestore }},
    }

    // Given
    for _, test := range tests {
        t.Setenv(test.envName, test.envValue)
    }

    // When
    _ = parseFlags([]string{}, log.NewEntry(log.New()))

    // Then
    for _, test := range tests {
        t.Run(
            test.envName, func(t *testing.T) {
                require.Equal(t, test.expected, test.actual())
            },
        )
    }
}

func Test_ParseFlag(t *testing.T) {
    tests := []struct {
        name     string
        value    string
        expected interface{}
        actual   func() interface{}
    }{
        {flagAddrKey, ":1111", ":1111", func() interface{} { return flagAddr }},
        {flagStoreIntKey, "120", uint(120), func() interface{} { return flagStoreInt }},
        {flagStorePathKey, "/test", "/test", func() interface{} { return flagStorePath }},
        {flagRestoreKey, "false", false, func() interface{} { return flagRestore }},
    }

    // Given
    var args []string

    for _, test := range tests {
        args = append(args, fmt.Sprintf("-%s=%s", test.name, test.value))
    }

    // When
    _ = parseFlags(args, log.NewEntry(log.New()))

    // Then
    for _, test := range tests {
        t.Run(
            fmt.Sprintf("Flag %s", test.name), func(t *testing.T) {
                require.Equal(t, test.expected, test.actual())
            },
        )
    }
}

func Test_PassNoFlag_SetDefault(t *testing.T) {
    tests := []struct {
        name     string
        expected interface{}
        actual   func() interface{}
    }{
        {flagAddrKey, flagAddrDefault, func() interface{} { return flagAddr }},
        {flagStoreIntKey, flagStoreIntDefault, func() interface{} { return flagStoreInt }},
        {flagStorePathKey, flagStorePathDefault, func() interface{} { return flagStorePath }},
        {flagRestoreKey, flagRestoreDefault, func() interface{} { return flagRestore }},
    }

    // When
    _ = parseFlags([]string{}, log.NewEntry(log.New()))

    // Then
    for _, test := range tests {
        t.Run(
            fmt.Sprintf("Flag %s", test.name), func(t *testing.T) {
                require.Equal(t, test.expected, test.actual())
            },
        )
    }
}
