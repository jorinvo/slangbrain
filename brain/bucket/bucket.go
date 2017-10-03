package bucket

// id is a chat id as int64
// time is an unix timestamp in seconds as int64
// phrase is a bucket sequence as uint64
// scoreupdate is an int64
// newscore is an int64
var (
	// Modes maps id -> Mode.
	Modes = []byte("modes")
	// id+phrase -> gob(Phrase)
	Phrases = []byte("phrases")
	// id+phrase -> time
	Studytimes = []byte("studytimes")
	// id+phrase -> time
	PhraseAddTimes = []byte("phraseaddtimes")
	// id -> phrase+phrase+phrase+...
	NewPhrases = []byte("newphrases")
	// id -> time
	Reads = []byte("reads")
	// Activities maps id -> time.
	Activities = []byte("activities")
	// Subscriptions maps id -> '1'.
	Subscriptions = []byte("subscriptions")
	// Profiles maps id -> gob(profile).
	Profiles = []byte("profiles")
	// RegisterDates maps id -> time.
	RegisterDates = []byte("registerdates")
	// Stattimes maps id -> time.
	Stattimes = []byte("stattimes")
	// Scoretotals maps id -> int64.
	Scoretotals = []byte("scoretotals")
	// Zeroscores maps id -> int64.
	Zeroscores = []byte("zeroscores")
	// Studies maps id+time -> phrase+scoreupdate+newscore.
	Studies = []byte("studies")
	// MessageIDs maps string -> time.
	MessageIDs = []byte("messageids")
	// AuthTokens maps token -> id.
	AuthTokens = []byte("authtokens")
	// AuthUsers maps id -> time+string(token).
	AuthUsers = []byte("authusers")
	// PendingImports maps id -> gob([]Phrase).
	PendingImports = []byte("pendingimports")
	// PrevPayloads maps id -> time+string(payload).
	PrevPayloads = []byte("prevpayloads")
	// Imports maps id -> int64.
	Imports = []byte("imports")
	// Notifies maps id -> int64.
	Notifies = []byte("notifies")
)

// All is a list of all bucket names.
var All = [][]byte{
	Modes,
	Phrases,
	Studytimes,
	PhraseAddTimes,
	NewPhrases,
	Reads,
	Activities,
	Subscriptions,
	Profiles,
	RegisterDates,
	Stattimes,
	Scoretotals,
	Zeroscores,
	Studies,
	MessageIDs,
	AuthTokens,
	AuthUsers,
	PendingImports,
	PrevPayloads,
	Imports,
	Notifies,
}
