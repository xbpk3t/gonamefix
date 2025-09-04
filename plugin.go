package gonamefix

import (
	"golang.org/x/tools/go/analysis"
)

// Plugin is the exported plugin variable for golangci-lint
var Plugin = &GonameFixPlugin{}

// GonameFixPlugin implements the golangci-lint plugin interface
type GonameFixPlugin struct{}

// GetLinter returns the linter for golangci-lint module plugin system
func (p *GonameFixPlugin) GetLinter() *analysis.Analyzer {
	return Analyzer
}

// GetAnalyzers returns all analyzers provided by this plugin
func (p *GonameFixPlugin) GetAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{Analyzer}
}

// GetLinters returns all analyzers provided by this plugin (alternative interface)
func (p *GonameFixPlugin) GetLinters() []*analysis.Analyzer {
	return []*analysis.Analyzer{Analyzer}
}

// New returns a new instance of the linter
func New() *analysis.Analyzer {
	return Analyzer
}

// GetLinter returns the linter for golangci-lint module plugin system (function version)
func GetLinter() *analysis.Analyzer {
	return Analyzer
}

// GetAnalyzers returns all analyzers provided by this plugin (function version)
func GetAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{Analyzer}
}
