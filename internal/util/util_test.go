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
			[]string{"HOME", "PATH", "PWD"},
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

func TestCreateTempFile(t *testing.T) {
	tests := []struct {
		name          string
		prefix        string
		content       string
		expectedError string
	}{
		{
			name:          "NoContent",
			prefix:        "test-",
			content:       "",
			expectedError: "",
		},
		{
			name:          "WithContent",
			prefix:        "test-",
			content:       "something",
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			filepath, remove, err := CreateTempFile(tc.prefix, tc.content)
			defer remove()

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Empty(t, filepath)
				assert.Nil(t, remove)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, filepath)
				assert.NotNil(t, remove)
			}
		})
	}
}
