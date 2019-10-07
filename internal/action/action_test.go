package action

import (
	"context"
	"testing"

	"github.com/moorara/cherry/internal/spec"
	"github.com/stretchr/testify/assert"
)

func TestContextWithSpec(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		spec spec.Spec
	}{
		{
			name: "OK",
			ctx:  context.Background(),
			spec: spec.Spec{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := ContextWithSpec(tc.ctx, tc.spec)

			s, ok := ctx.Value(specKey).(spec.Spec)
			assert.True(t, ok)
			assert.Equal(t, tc.spec, s)
		})
	}
}

func TestSpecFromContext(t *testing.T) {
	tests := []struct {
		name         string
		ctx          context.Context
		expectedSpec spec.Spec
	}{
		{
			name:         "WithoutSpec",
			ctx:          context.Background(),
			expectedSpec: spec.Spec{},
		},
		{
			name: "WithSpec",
			ctx: context.WithValue(context.Background(), specKey, spec.Spec{
				ToolName: "cherry",
			}),
			expectedSpec: spec.Spec{
				ToolName: "cherry",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := SpecFromContext(tc.ctx)
			assert.Equal(t, tc.expectedSpec, s)
		})
	}
}
