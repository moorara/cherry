package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTest(t *testing.T) {
	tests := []struct {
		name string
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotNil(t, tc)
		})
	}
}

func TestTestSynopsis(t *testing.T) {
	cmd := &Test{}
	synopsis := cmd.Synopsis()
	assert.Equal(t, testSynopsis, synopsis)
}

func TestTestHelp(t *testing.T) {
	cmd := &Test{}
	help := cmd.Help()
	assert.Equal(t, testHelp, help)
}

func TestTestRun(t *testing.T) {
	tests := []struct {
		name string
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotNil(t, tc)
		})
	}
}
