package formula

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBuildInfo(t *testing.T) {
	tests := []struct {
		name    string
		formula *formula
		ctx     context.Context
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotNil(t, tc)
		})
	}
}

func TestGetLDFlags(t *testing.T) {
	tests := []struct {
		name    string
		formula *formula
		ctx     context.Context
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotNil(t, tc)
		})
	}
}

func TestCompile(t *testing.T) {
	tests := []struct {
		name    string
		formula *formula
		ctx     context.Context
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotNil(t, tc)
		})
	}
}

func TestCrossCompile(t *testing.T) {
	tests := []struct {
		name    string
		formula *formula
		ctx     context.Context
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotNil(t, tc)
		})
	}
}
