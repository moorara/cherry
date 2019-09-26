package step

// Step is an atomic piece of functionality that can be reverted.
type Step interface {
	Dry() error
	Run() error
	Revert() error
}
