package step

import "context"

// Step is an atomic piece of functionality that can be reverted.
type Step interface {
	Dry(context.Context) error
	Run(context.Context) error
	Revert(context.Context) error
}
