package nmt_test

import (
	"testing"

	"github.com/lucky-peak/nmtscope/server/nmt"
)

func TestGenerateNMTReport(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		jcmd      string
		pid       int
		reportDir string
	}{
		// TODO: Add test cases.
		{
			name:      "Generate NMT report",
			jcmd:      "jcmd",
			pid:       53732,
			reportDir: "/tmp/nmt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nmt.GenerateNMTReport(tt.jcmd, tt.pid, tt.reportDir)
		})
	}
}
