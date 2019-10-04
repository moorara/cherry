package step

import (
	"context"
)

type mockStep struct {
	DryInCtx       context.Context
	DryOutError    error
	RunInCtx       context.Context
	RunOutError    error
	RevertInCtx    context.Context
	RevertOutError error
}

func (m *mockStep) Dry(ctx context.Context) error {
	m.DryInCtx = ctx
	return m.DryOutError
}

func (m *mockStep) Run(ctx context.Context) error {
	m.RunInCtx = ctx
	return m.RunOutError
}

func (m *mockStep) Revert(ctx context.Context) error {
	m.RevertInCtx = ctx
	return m.RevertOutError
}
