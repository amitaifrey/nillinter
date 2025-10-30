package main

import (
	an "github.com/amitaifrey/nillinter/internal/analyzer"

	"golang.org/x/tools/go/analysis"
)

// New is the entrypoint expected by golangci-lint Go Plugin System.
// It may receive a config object (from linters.settings.custom.<name>.settings) but we don't use it.
func New(conf any) ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{an.Analyzer}, nil
}
