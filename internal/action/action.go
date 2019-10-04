package action

import (
	"context"
)

// Action is an ordered list of steps that can be reverted.
type Action interface {
	Dry(context.Context) error
	Run(context.Context) error
	Revert(context.Context) error
}
