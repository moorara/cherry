package step

import "context"

// Step is an atomic piece of functionality that can be reverted.
type Step interface {
	Dry(ctx context.Context) error
	Run(ctx context.Context) error
	Revert(ctx context.Context) error
}
