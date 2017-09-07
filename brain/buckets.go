package brain

// For docs see below
var (
	bucketModes          = []byte("modes")
	bucketPhrases        = []byte("phrases")
	bucketStudytimes     = []byte("studytimes")
	bucketPhraseAddTimes = []byte("phraseaddtimes")
	bucketNewPhrases     = []byte("newphrases")
	bucketReads          = []byte("reads")
	bucketActivities     = []byte("activities")
	bucketSubscriptions  = []byte("subscriptions")
	bucketProfiles       = []byte("profiles")
	bucketRegisterDates  = []byte("registerdates")
	bucketStattimes      = []byte("stattimes")
	bucketScoretotals    = []byte("scoretotals")
	bucketZeroscores     = []byte("zeroscores")
	bucketStudies        = []byte("studies")
	bucketMessageIDs     = []byte("messageids")
	bucketAuthTokens     = []byte("authtokens")
	bucketAuthUsers      = []byte("authusers")
	bucketPendingImports = []byte("pendingimports")
	bucketPrevPayloads   = []byte("prevpayloads")
)

// id is a chat id as int64
// time is an unix timestamp in seconds as int64
// phrase is a bucket sequence as uint64
// scoreupdate is an int64
// newscore is an int64
var allBuckets = [][]byte{
	// id -> Mode
	bucketModes,
	// id+phrase -> gob(Phrase)
	bucketPhrases,
	// id+phrase -> time
	bucketStudytimes,
	// id+phrase -> time
	bucketPhraseAddTimes,
	// id -> phrase+phrase+phrase+...
	bucketNewPhrases,
	// id -> time
	bucketReads,
	// id -> time
	bucketActivities,
	// id -> '1'
	bucketSubscriptions,
	// id -> gob(profile)
	bucketProfiles,
	// id -> time
	bucketRegisterDates,
	// id -> time
	bucketStattimes,
	// id -> int64
	bucketScoretotals,
	// id -> int64
	bucketZeroscores,
	// id+time -> phrase+scoreupdate+newscore
	bucketStudies,
	// string -> time
	bucketMessageIDs,
	// token -> id
	bucketAuthTokens,
	// id -> time+string(token)
	bucketAuthUsers,
	// id -> gob([]Phrase)
	bucketPendingImports,
	// id -> time+string(payload)
	bucketPrevPayloads,
}
