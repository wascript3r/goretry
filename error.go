package goretry

type State uint32

const (
	ContinueState State = iota
	InformationalState
	StopState
	ExitState
	InvalidState
)

func IsValidState(state State) bool {
	switch state {
	case ContinueState, InformationalState, StopState, ExitState:
		return true
	}
	return false
}

type Error struct {
	State    State
	Original error
}

func NewError(state State, err error) *Error {
	return &Error{state, err}
}

func Continue(err error) *Error {
	return NewError(ContinueState, err)
}
