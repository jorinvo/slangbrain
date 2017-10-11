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
	// ErrNotReady signals that the requested data is not ready.
	ErrNotReady = errors.New("not ready")
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
	Phrase      string `json:"phrase,omitempty"`
	Explanation string `json:"explanation,omitempty"`
	Score       int    `json:"score,omitempty"`
}

// Stats describes statistics for a single user.
type Stats struct {
	// Added is the number of phrases added in the last interval.
	Added int
	// Studied is the number of phrases studied in the last interval.
	Studied int
	// Score is the total score of all phrases of a user.
	Score int
	// Rank is the rank by score of a user compared to all other users.
	Rank int
}

// Profile abstracts a user profile.
// It is only used for reading information.
// Can be read from remote or from cache.
type Profile interface {
	Name() string
	Locale() string
	Timezone() float64
}
