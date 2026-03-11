package osexit_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"link-shortener/cmd/staticlint/osexit"
)

func TestOsExitAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), osexit.Analyzer, "./...")
}
