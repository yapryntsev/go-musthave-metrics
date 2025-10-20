package main

import (
    "fmt"
    "github.com/stretchr/testify/require"
    "go.uber.org/zap/zaptest"
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
        {envPollIntKey, "10", uint(10), func() interface{} { return flagPollInt }},
        {envReportIntKey, "30", uint(30), func() interface{} { return flagReportInt }},
    }

    // Given
    for _, test := range tests {
        t.Setenv(test.envName, test.envValue)
    }

    // When
    parseFlags([]string{}, zaptest.NewLogger(t))

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
        {flagPollIntKey, "90", uint(90), func() interface{} { return flagPollInt }},
        {flagReportIntKey, "20", uint(20), func() interface{} { return flagReportInt }},
    }

    // Given
    var args []string

    for _, test := range tests {
        args = append(args, fmt.Sprintf("-%s=%s", test.name, test.value))
    }

    // When
    parseFlags(args, zaptest.NewLogger(t))

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
        {flagPollIntKey, flagPollIntDefault, func() interface{} { return flagPollInt }},
        {flagReportIntKey, flagReportIntDefault, func() interface{} { return flagReportInt }},
    }

    // When
    parseFlags([]string{}, zaptest.NewLogger(t))

    // Then
    for _, test := range tests {
        t.Run(
            fmt.Sprintf("Flag %s", test.name), func(t *testing.T) {
                require.Equal(t, test.expected, test.actual())
            },
        )
    }
}

func Test_PassInvalidEnv_SetDefault(t *testing.T) {
    tests := []struct {
        envName  string
        envValue string
        expected interface{}
        actual   func() interface{}
    }{
        {envPollIntKey, "hi", flagPollIntDefault, func() interface{} { return flagPollInt }},
        {envReportIntKey, "ho", flagReportIntDefault, func() interface{} { return flagReportInt }},
    }

    // Given
    for _, test := range tests {
        t.Setenv(test.envName, test.envValue)
    }

    // When
    parseFlags([]string{}, zaptest.NewLogger(t))

    // Then
    for _, test := range tests {
        t.Run(
            test.envName, func(t *testing.T) {
                require.Equal(t, test.expected, test.actual())
            },
        )
    }
}
