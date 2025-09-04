package gonamefix

import (
	"golang.org/x/tools/go/analysis"
)

// LinterPlugin is the plugin implementation for golangci-lint
type LinterPlugin struct{}

// GetLinters returns the linters provided by this plugin
func (p *LinterPlugin) GetLinters() []*analysis.Analyzer {
	return []*analysis.Analyzer{Analyzer}
}

// New creates a new instance of the plugin
func New() *LinterPlugin {
	return &LinterPlugin{}
}
