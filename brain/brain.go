package brain

import "time"

// Mode is the state of a chat.
// We need to keep track of the state each chat is in.
type Mode int

const (
	// ModeIdle ...
	ModeIdle Mode = iota
	// ModeAdd ...
	ModeAdd
	// ModeStudy ...
	ModeStudy
	// ModeGetStarted ...
	ModeGetStarted
)

// Study ...
type Study struct {
	Phrase      string
	Explanation string
	Total       int
	Next        time.Duration
}

// Score ...
type Score int

const (
	// ScoreBad ...
	ScoreBad = iota - 1
	// ScoreOK ...
	ScoreOK
	// ScoreGood ...
	ScoreGood
)

// Phrase ...
type Phrase struct {
	Phrase      string
	Explanation string
	Score       Score
}

func newPhrase(phrase, explanation string) Phrase {
	return Phrase{
		Phrase:      phrase,
		Explanation: explanation,
	}
}
