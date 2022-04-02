package main

import (
	"github.com/farislr/commoneer/analyzer"
	"golang.org/x/tools/go/analysis"
)

type analyzerPlugin struct{}

func (*analyzerPlugin) GetAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		analyzer.QueryxAnalyzer,
	}
}

var AnalyzerPlugin analyzerPlugin
