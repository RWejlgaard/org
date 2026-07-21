package model

// TodoState represents the state of a todo item
type TodoState string

const (
	// StateNone is the empty (no TODO keyword) state. Concrete TODO states are
	// user-configurable (see config.States), so they are not hardcoded here.
	StateNone TodoState = ""
)
