package brain

import "time"

const (
	// Time in minutes
	// When study times are updated they are randomly placed
	// somewhere between the new time and new time + studyTimeDiffusion
	// to mix up the order in which words are studied.
	studyTimeDiffusion = 30
	// Maximum number of new studies at a time
	maxNewStudies = 20
	// Minimum number of studies needed to be due before notifying user
	dueMinCount = 10
	// Time user has to be inactive before being notified
	dueMinInactive = 1 * time.Hour
	// Cache profiles for one month
	profileMaxCacheTime = 30 * 24 * time.Hour
	// Number of chars a token gets
	authTokenLength = 77
	// Tokens are valid for one month
	authTokenMaxAge = 30 * 24 * time.Hour
	// Handle same payload only once in the given interval to prevent accidentally sending payloads twice
	payloadDuplicateInterval = 10 * time.Second
	// Time after which message IDs are cleared, adjust to keep bucket size from exploding
	messageIDmaxAge = 24 * time.Hour
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
