package main

import (
	"github.com/farislr/commoneer/analyzer"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(
		analyzer.QueryxAnalyzer,
	)
}
