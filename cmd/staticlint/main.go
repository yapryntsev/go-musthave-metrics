package main

import (
	"github.com/yapryntsev/go-musthave-metrics/cmd/staticlint/mainexit"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck/st1000"
	"honnef.co/go/tools/stylecheck/st1003"
)

func main() {
	var checks []*analysis.Analyzer
	for _, check := range staticcheck.Analyzers {
		checks = append(checks, check.Analyzer)
	}

	checks = append(
		checks,
		assign.Analyzer,
		buildtag.Analyzer,
		copylock.Analyzer,
		defers.Analyzer,
		nilfunc.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		usesgenerics.Analyzer,
		st1000.Analyzer,
		st1003.Analyzer,
		mainexit.Analyzer,
	)

	multichecker.Main(checks...)
}
