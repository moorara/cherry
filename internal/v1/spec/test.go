package spec

import (
	"flag"
)

const (
	defaultCoverMode  = "set"
	defaultReportPath = "coverage"
)

type (
	// Test specifies the configurations for test command
	Test struct {
		CoverMode  string `json:"covermode" yaml:"covermode"`
		ReportPath string `json:"reportPath" yaml:"report_path"`
	}
)

// SetDefaults set default values for empty fields
func (t *Test) SetDefaults() {
	if t.CoverMode == "" {
		t.CoverMode = defaultCoverMode
	}

	if t.ReportPath == "" {
		t.ReportPath = defaultReportPath
	}
}

// FlagSet returns a flag set for parsing input arguments
func (t *Test) FlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.StringVar(&t.CoverMode, "covermode", t.CoverMode, "")
	fs.StringVar(&t.ReportPath, "report-path", t.ReportPath, "")

	return fs
}
