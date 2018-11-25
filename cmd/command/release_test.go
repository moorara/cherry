package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRelease(t *testing.T) {
	tests := []struct {
		name string
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotNil(t, tc)
		})
	}
}

func TestReleaseSynopsis(t *testing.T) {
	cmd := &Release{}
	synopsis := cmd.Synopsis()
	assert.Equal(t, releaseSynopsis, synopsis)
}

func TestReleaseHelp(t *testing.T) {
	cmd := &Release{}
	help := cmd.Help()
	assert.Equal(t, releaseHelp, help)
}

func TestReleaseRun(t *testing.T) {
	tests := []struct {
		name string
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotNil(t, tc)
		})
	}
}
