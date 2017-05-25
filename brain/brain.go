package brain

// Mode is the state of a chat.
// We need to keep track of the state each chat is in.
type Mode int

const (
	// ModeIdle ...
	// Currently not in use
	ModeIdle Mode = iota
	// ModeAdd ...
	ModeAdd
	// ModeStudy ...
	ModeStudy
)

// Studymode ...
type Studymode int

const (
	// ButtonsExplanation  ...
	ButtonsExplanation Studymode = 1 << iota
	// ButtonsPhrase ...
	ButtonsPhrase
	// TypeExplanation ...
	TypeExplanation
	// TypePhrase ...
	TypePhrase
)

// Studymodes ...
var Studymodes = []Studymode{ButtonsExplanation, ButtonsPhrase, TypeExplanation, TypePhrase}

// Study ...
type Study struct {
	ID          int
	Mode        Studymode
	Phrase      string
	Explanation string
	Total       int
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
