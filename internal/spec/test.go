package spec

import (
	"flag"
)

const (
	defaultReportPath = "coverage"
)

type (
	// Test specifies the configurations for test command
	Test struct {
		ReportPath string `json:"reportPath" yaml:"report_path"`
	}
)

// SetDefaults set default values for empty fields
func (t *Test) SetDefaults() {
	if t.ReportPath == "" {
		t.ReportPath = defaultReportPath
	}
}

// FlagSet returns a flag set for parsing input arguments
func (t *Test) FlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.StringVar(&t.ReportPath, "report-path", t.ReportPath, "")

	return fs
}
