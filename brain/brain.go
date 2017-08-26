// Package brain handles all business logic.
// It handles data storage, retrieving and updating.
// It's independent from the used bot platform and user interaction.
package brain

import (
	"errors"
	"time"
)

var (
	// ErrExists signals that the thing to be added has been added already.
	ErrExists = errors.New("already exists")
	// ErrNotFound signals that the requested entry couldn't be found.
	ErrNotFound = errors.New("not found")
)

// Mode is the state of a chat.
// We need to keep track of the state each chat is in.
type Mode int

const (
	// ModeMenu shows the main menu.
	ModeMenu Mode = iota
	// ModeAdd lets the user add new phrases.
	ModeAdd
	// ModeStudy goes to phrases ready to study.
	ModeStudy
	// ModeGetStarted sends an introduction to the user.
	ModeGetStarted
	// ModeFeedback allows the user to send a message that is ready by a human.
	ModeFeedback
)

// Study is a study the current study the user needs to answer.
type Study struct {
	// Phrase is the phrase the user needs to guess.
	Phrase string
	// Explanation is the explanation displayed to the user.
	Explanation string
	// Total is the total number of studies ready, including the current one.
	Total int
	// Next contains the time until the next study is available;
	// it's only set if Total is 0.
	Next time.Duration
}

// Phrase describes a phrase the user saved.
type Phrase struct {
	Phrase      string
	Explanation string
	Score       int
}
