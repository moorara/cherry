package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	tests := []struct {
		name             string
		expectedLogLevel string
	}{
		{
			name:             "Defauts",
			expectedLogLevel: defaultLogLevel,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedLogLevel, Config.LogLevel)
		})
	}
}
