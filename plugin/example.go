package plugin

import (
	"github.com/amitaifrey/nillinter/internal/analyzer"
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

// nillinterPlugin implements the LinterPlugin interface.
type nillinterPlugin struct{}

func (p *nillinterPlugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{analyzer.Analyzer}, nil
}

func (p *nillinterPlugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}

func init() {
	register.Plugin("nillinter", New)
}

// New is the entrypoint expected by golangci-lint Go Plugin System.
// It may receive a config object (from linters.settings.custom.<name>.settings) but we don't use it.
func New(conf any) (register.LinterPlugin, error) {
	return &nillinterPlugin{}, nil
}
