package brain

// Store is the interface to an underlying database.
// Keeping it separated allows to switch the database implementation.
type Store interface {
	UseChat(int64, int, string) (Chat, error)
	AddFact(int64, string) error
	FindFacts(int, string) ([]Fact, error)
	SetMode(int64, Mode) error
}
