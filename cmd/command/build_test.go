package command

import (
	"testing"

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
	cmd := &Build{}
	help := cmd.Help()
	assert.Equal(t, buildHelp, help)
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
