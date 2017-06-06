package brain

import "time"

const (
	// Time to wait for first study in hours
	firstStudytime = 3
	// base time in hours to use to calculate next study time
	baseStudytime = 10
	// Time in minutes
	// When study times are updated they are randomly placed
	// somewhere between the new time and new time + studyTimeDiffusion
	// to mix up the order in which words are studied.
	studyTimeDiffusion = 30
	// Maximum number of new studies per day
	newPerDay = 10
	// Minimum number of studies needed to be due before notifying user
	dueMinCount = 5
	// Time user has to be inactive before being notified
	dueMinInactive = 10 * time.Minute
)

var (
	bucketModes         = []byte("modes")
	bucketPhrases       = []byte("phrases")
	bucketStudytimes    = []byte("studytimes")
	bucketReads         = []byte("reads")
	bucketActivities    = []byte("activities")
	bucketSubscriptions = []byte("subscriptions")
)
var buckets = [][]byte{
	bucketModes,
	bucketPhrases,
	bucketStudytimes,
	bucketReads,
	bucketActivities,
	bucketSubscriptions,
}

// Mode is the state of a chat.
// We need to keep track of the state each chat is in.
type Mode int

const (
	// ModeMenu ...
	ModeMenu Mode = iota
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
