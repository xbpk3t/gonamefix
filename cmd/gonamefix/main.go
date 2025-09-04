package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/xbpk3t/gonamefix"
)

func main() {
	// Create analyzer with test configuration
	config := gonamefix.Config{
		Check: [][]string{
			{"request", "req"},
			{"response", "res"},
			{"configuration", "config"},
		},
		ExcludeFiles:  []string{"*.pb.go", "*_test.go"},
		ExcludeDirs:   []string{"vendor", "node_modules", ".git"},
		CaseSensitive: false,
	}
	
	analyzer := gonamefix.NewAnalyzer(config)
	singlechecker.Main(analyzer)
}
