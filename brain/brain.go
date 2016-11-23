package brain

// Mode is the state of a chat.
// We need to keep track of the state each chat is in.
type Mode int

const (
	// IdleMode is the default.
	IdleMode = iota
	// AddMode creates a new fact for what a user sends next.
	AddMode
)

// Fact is one element a user saved.
// It also matches a row in the facts table.
//
// Currently a fact just contains a single text content.
// In the future there might be different fields with different functions.
//
// A fact belongs to a Chat.
// This way users can share knowledge using a group chat
// or a single user can create multiple groups as "studying groups" to separate different kinds of facts.
type Fact struct {
	Content string
	Chat    Chat
}

// Chat can be shared by multiple users.
// We need to keep track of Title and Mode.
type Chat struct {
	ID    int64
	Title string
	Mode
}
