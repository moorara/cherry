package action

import (
	"context"

	"github.com/moorara/cherry/internal/spec"
)

// contextKey is the type for the keys added to context.
type contextKey string

const specKey = contextKey("Spec")

// ContextWithSpec returns a new context that has a copy of spec.
func ContextWithSpec(ctx context.Context, s spec.Spec) context.Context {
	return context.WithValue(ctx, specKey, s)
}

// SpecFromContext returns a copy of spec from context.
// If no spec found, a default (zero) spec will be returned.
func SpecFromContext(ctx context.Context) spec.Spec {
	s, ok := ctx.Value(specKey).(spec.Spec)
	if ok {
		return s
	}

	return spec.Spec{}
}

// Action is an ordered list of steps that can be reverted.
type Action interface {
	Dry(context.Context) error
	Run(context.Context) error
	Revert(context.Context) error
}
