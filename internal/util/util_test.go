package util

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnsureCommands(t *testing.T) {
	tests := []struct {
		name          string
		commands      []string
		expectedError error
	}{
		{
			"NoCommand",
			[]string{},
			nil,
		},
		{
			"SingleCommand",
			[]string{"cut"},
			nil,
		},
		{
			"MultipleCommand",
			[]string{"cut", "date", "grep"},
			nil,
		},
		{
			"Error",
			[]string{"unknown"},
			errors.New("unknown command is no available"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := EnsureCommands(tc.commands...)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestEnsureEnvVars(t *testing.T) {
	tests := []struct {
		name          string
		variables     []string
		expectedError error
	}{
		{
			"NoEnvVar",
			[]string{},
			nil,
		},
		{
			"SingleEnvVar",
			[]string{"HOME"},
			nil,
		},
		{
			"MultipleEnvVar",
			[]string{"HOME", "SHELL", "USER"},
			nil,
		},
		{
			"Error",
			[]string{"UNKNOWN"},
			errors.New("UNKNOWN environment variable is not set"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := EnsureEnvVars(tc.variables...)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}
