package brain

import "time"

const (
	// When study times are updated they are randomly placed
	// somewhere between the new time and new time + studyTimeDiffusion*studyIntervals[i]
	// to mix up the order in which words are studied.
	studyTimeDiffusion = 0.2
	// Maximum number of new studies at a time
	maxNewStudies = 20
	// Delay scheduled new phrases are placed at
	newStudyDelay = 24 * time.Hour
	// Minimum number of studies needed to be due before notifying user
	dueMinCount = 10
	// Time user has to be inactive before being notified
	dueMinInactive = 1 * time.Hour
	// Cache profiles locally
	profileMaxCacheTime = 3 * 24 * time.Hour
	// Number of chars a token gets
	authTokenLength = 77
	// Handle same payload only once in the given interval to prevent accidentally sending payloads twice
	payloadDuplicateInterval = 5 * time.Second
	// Time after which message IDs are cleared, adjust to keep bucket size from exploding
	messageIDmaxAge = 24 * time.Hour
	// Hour of the day after which no notifications can be sent to the user
	nightStart = 21
	// Hour of the day from which on notifications can be sent to the user
	nightEnd = 7
	// Show user stats once a week
	statInterval = 7 * 24 * time.Hour
)

var studyIntervals = [14]time.Duration{
	time.Hour,
	8 * time.Hour,
	20 * time.Hour,
	44 * time.Hour,
	(4*24 - 2) * time.Hour,
	(7*24 - 2) * time.Hour,
	(14*24 - 2) * time.Hour,
	(30*24 - 2) * time.Hour,
	(60*24 - 2) * time.Hour,
	(100*24 - 2) * time.Hour,
	(5*30*24 - 2) * time.Hour,
	(8*30*24 - 2) * time.Hour,
	(12*30*24 - 2) * time.Hour,
	(15*30*24 - 2) * time.Hour,
}
