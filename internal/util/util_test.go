package util

import (
	"errors"
	"os"
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

func TestSetEnvVars(t *testing.T) {
	tests := []struct {
		name          string
		keyVals       []string
		expectedError string
	}{
		{
			name:          "NoKeyValue",
			keyVals:       []string{},
			expectedError: "",
		},
		{
			name:          "MismatchingKeyValues",
			keyVals:       []string{"TEST_FOO"},
			expectedError: "mismatching key-value pairs",
		},
		{
			name:          "InvalidKeyValues",
			keyVals:       []string{"", "foo"},
			expectedError: "setenv: invalid argument",
		},
		{
			name: "Success",
			keyVals: []string{
				"TEST_FOO", "foo",
				"TEST_BAR", "bar",
			},
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reset, err := SetEnvVars(tc.keyVals...)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Nil(t, reset)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, reset)

				for i := 0; i < len(tc.keyVals); i += 2 {
					assert.Equal(t, tc.keyVals[i+1], os.Getenv(tc.keyVals[i]))
				}

				err := reset()
				assert.NoError(t, err)

				for i := 0; i < len(tc.keyVals); i += 2 {
					assert.Equal(t, "", os.Getenv(tc.keyVals[i]))
				}
			}
		})
	}
}
