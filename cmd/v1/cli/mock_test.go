package cli

import (
	"context"

	"github.com/moorara/cherry/internal/v1/formula"
)

type mockUI struct {
	OutputInMessage string

	InfoInMessage string

	WarnInMessage string

	ErrorInMessage string

	AskInQuery   string
	AskOutResult string
	AskOutError  error

	AskSecretInQuery   string
	AskSecretOutResult string
	AskSecretOutError  error
}

func (m *mockUI) Output(message string) {
	m.OutputInMessage = message
}

func (m *mockUI) Info(message string) {
	m.InfoInMessage = message
}

func (m *mockUI) Warn(message string) {
	m.WarnInMessage = message
}

func (m *mockUI) Error(message string) {
	m.ErrorInMessage = message
}

func (m *mockUI) Ask(query string) (string, error) {
	m.AskInQuery = query
	return m.AskOutResult, m.AskOutError
}

func (m *mockUI) AskSecret(query string) (string, error) {
	m.AskSecretInQuery = query
	return m.AskSecretOutResult, m.AskSecretOutError
}

type mockFormula struct {
	InfoInMessage string

	WarnInMessage string

	ErrorInMessage string

	CoverInCtx    context.Context
	CoverOutError error

	CompileInCtx    context.Context
	CompileOutError error

	CrossCompileInCtx     context.Context
	CrossCompileOutResult []string
	CrossCompileOutError  error

	ReleaseInCtx     context.Context
	ReleaseInLevel   formula.ReleaseLevel
	ReleaseInComment string
	ReleaseOutError  error
}

func (m *mockFormula) Info(message string) {
	m.InfoInMessage = message
}

func (m *mockFormula) Warn(message string) {
	m.WarnInMessage = message
}

func (m *mockFormula) Error(message string) {
	m.ErrorInMessage = message
}

func (m *mockFormula) Cover(ctx context.Context) error {
	m.CoverInCtx = ctx
	return m.CoverOutError
}

func (m *mockFormula) Compile(ctx context.Context) error {
	m.CompileInCtx = ctx
	return m.CompileOutError
}

func (m *mockFormula) CrossCompile(ctx context.Context) ([]string, error) {
	m.CrossCompileInCtx = ctx
	return m.CrossCompileOutResult, m.CrossCompileOutError
}

func (m *mockFormula) Release(ctx context.Context, level formula.ReleaseLevel, comment string) error {
	m.ReleaseInCtx = ctx
	m.ReleaseInLevel = level
	m.ReleaseInComment = comment
	return m.ReleaseOutError
}
