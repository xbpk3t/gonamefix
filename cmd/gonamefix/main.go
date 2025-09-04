package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/xbpk3t/gonamefix"
)

func main() {
	singlechecker.Main(gonamefix.Analyzer)
}
