package build

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBuilder(t *testing.T) {
	tests := []struct {
		name string
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotNil(t, tc)
		})
	}
}
