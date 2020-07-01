package goretry

import "fmt"

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

func (s State) String() string {
	switch s {
	case ContinueState:
		return "Continue"
	case InformationalState:
		return "Informational"
	case StopState:
		return "Stop"
	case ExitState:
		return "Exit"
	default:
		return "Invalid"
	}
}

type Error struct {
	State    State
	Original error
}

func NewError(state State, err error) *Error {
	return &Error{state, err}
}

func (e *Error) String() string {
	return fmt.Sprintf("%s (state: %s)", e.Original, e.State)
}

func Continue(err error) *Error {
	return NewError(ContinueState, err)
}

func Info(err error) *Error {
	return NewError(InformationalState, err)
}

func Stop(err error) *Error {
	return NewError(StopState, err)
}

func Exit(err error) *Error {
	return NewError(ExitState, err)
}
