package command

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/moorara/cherry/internal/spec"
	"github.com/stretchr/testify/assert"
)

func TestNewBuild(t *testing.T) {
	tests := []struct {
		name string
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotNil(t, tc)
		})
	}
}

func TestBuildSynopsis(t *testing.T) {
	cmd := &Build{}

	synopsis := cmd.Synopsis()
	assert.Equal(t, buildSynopsis, synopsis)
}

func TestBuildHelp(t *testing.T) {
	tests := []struct {
		spec spec.Spec
	}{
		{
			spec: spec.Spec{
				Build: spec.Build{
					BinaryFile: "bin/cherry",
				},
			},
		},
	}

	for _, tc := range tests {
		cmd := &Build{
			Spec: tc.spec,
		}

		var buf bytes.Buffer
		tmpl := template.Must(template.New("help").Parse(buildHelp))
		tmpl.Execute(&buf, cmd)
		expectedHelp := buf.String()

		help := cmd.Help()
		assert.Equal(t, expectedHelp, help)
	}
}

func TestBuildRun(t *testing.T) {
	tests := []struct {
		name string
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotNil(t, tc)
		})
	}
}
