package action

// Action is an ordered list of steps that can be reverted.
type Action interface {
	Dry() error
	Run() error
	Revert() error
}
