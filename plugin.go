package gonamefix

import (
	"golang.org/x/tools/go/analysis"
)

// GetLinter returns the linter for golangci-lint module plugin system
func GetLinter() *analysis.Analyzer {
	return Analyzer
}

// New returns a new instance of the linter
func New() *analysis.Analyzer {
	return Analyzer
}

// GetAnalyzers returns all analyzers provided by this plugin
func GetAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{Analyzer}
}
