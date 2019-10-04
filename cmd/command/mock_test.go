package command

import "context"

type mockCUI struct {
	OutputfInFormat string
	OutputfOutVals  []interface{}
	InfofInFormat   string
	InfofOutVals    []interface{}
	WarnfInFormat   string
	WarnfOutVals    []interface{}
	ErrorfInFormat  string
	ErrorfOutVals   []interface{}
}

func (m *mockCUI) Outputf(format string, vals ...interface{}) {
	m.OutputfInFormat = format
	m.OutputfOutVals = vals
}

func (m *mockCUI) Infof(format string, vals ...interface{}) {
	m.InfofInFormat = format
	m.InfofOutVals = vals
}

func (m *mockCUI) Warnf(format string, vals ...interface{}) {
	m.WarnfInFormat = format
	m.WarnfOutVals = vals
}

func (m *mockCUI) Errorf(format string, vals ...interface{}) {
	m.ErrorfInFormat = format
	m.ErrorfOutVals = vals
}

type mockAction struct {
	DryInCtx       context.Context
	DryOutError    error
	RunInCtx       context.Context
	RunOutError    error
	RevertInCtx    context.Context
	RevertOutError error
}

func (m *mockAction) Dry(ctx context.Context) error {
	m.DryInCtx = ctx
	return m.DryOutError
}

func (m *mockAction) Run(ctx context.Context) error {
	m.RunInCtx = ctx
	return m.RunOutError
}

func (m *mockAction) Revert(ctx context.Context) error {
	m.RevertInCtx = ctx
	return m.RevertOutError
}
