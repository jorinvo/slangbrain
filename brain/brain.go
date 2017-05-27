package brain

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

// Studymode ...
type Studymode int

const (
	// GuessPhrase  ...
	GuessPhrase Studymode = 1 << iota
	// GuessExplanation ...
	GuessExplanation
)

// Studymodes ...
var Studymodes = []Studymode{GuessPhrase, GuessExplanation}

// Study ...
type Study struct {
	Phrase      string
	Explanation string
	Mode        Studymode
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

// Phrase ...
type Phrase struct {
	Phrase           string
	Explanation      string
	ScorePhrase      Score
	ScoreExplanation Score
}

func newPhrase(phrase, explanation string) Phrase {
	return Phrase{
		Phrase:      phrase,
		Explanation: explanation,
	}
}
