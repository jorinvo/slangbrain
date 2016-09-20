package brain

import (
	"database/sql"

	"github.com/pkg/errors"
)

var (
	setupStmt = `
		CREATE TABLE IF NOT EXISTS facts (
			id INTEGER PRIMARY KEY,
			chatid BIGINT,
			content TEXT
		);
		CREATE TABLE IF NOT EXISTS chats (
			id INTEGER PRIMARY KEY,
			userid INTEGER,
			chatid BIGINT,
			chatname TEXT
		);
		CREATE UNIQUE INDEX IF NOT EXISTS uniqidcombi ON chats(chatid, userid);
	`

	addChatStmt = "REPLACE INTO chats (userid, chatid, chatname) VALUES ($1, $2, $3)"

	addFactStmt = "INSERT INTO facts (chatid, content) VALUES ($1, $2)"

	findFactsStmt = `
		SELECT id, facts.chatid, content, chatname
		FROM facts
		JOIN (SELECT chatid, chatname FROM chats WHERE userid = $1) c
		ON facts.chatid = c.chatid
		WHERE content LIKE '%' || $2 || '%'
	`
)

// Fact is one element a user saved.
// It also matches a row in the facts table.
//
// Currently a fact just contains a single text content.
// In the future there might be different fields with different functions.
//
// A fact belongs to a ChatID.
// This way users can share knowledge using a group chat
// or a single user can create multiple groups as "studying groups" to separate different kinds of facts.
type Fact struct {
	ID        int
	ChatID    int
	Content   string
	ChatTitle string
}

// Store is the interface to an underlying database.
// Keeping it separated allows to switch the database implementation.
type Store interface {
	AddChat(int, int64, string) error
	AddFact(int64, string) error
	FindFacts(int, string) ([]Fact, error)
}

// SQLStore implements a Store as classical SQL database.
// It uses the standard database/sql package and supports different backends.
// Current implementation is tested with SQLite.
type SQLStore struct {
	db *sql.DB
}

// CreateStore returns a new SQLStore with a database already setup.
// It forwards the arguments to the standard database/sql package.
func CreateStore(driverName, dataSourceName string) (SQLStore, error) {
	db, err := sql.Open(driverName, dataSourceName)
	store := SQLStore{db}
	if err != nil {
		return store, errors.Wrap(err, "failed to initialize database adapter")
	}

	_, err = db.Exec(setupStmt)
	if err != nil {
		return store, errors.Wrap(err, "failed to setup tables")
	}

	err = db.Ping()
	return store, errors.Wrap(err, "failed to connect to database")
}

// AddChat saves a relation between a userID and chatID.
// Relations are stored unique - even after repeated calling of AddChat.
func (store SQLStore) AddChat(userID int, chatID int64, chatTitle string) error {
	_, err := store.db.Exec(addChatStmt, userID, chatID, chatTitle)
	return errors.Wrapf(err, "failed to add chat for userID %d and chatID %d", userID, chatID)
}

// AddFact stores a new fact.
func (store SQLStore) AddFact(chatID int64, text string) error {
	_, err := store.db.Exec(addFactStmt, chatID, text)
	return errors.Wrapf(err, "failed to add fact for chatID %d: %s", chatID, text)
}

// FindFacts returns all facts that contain the query string.
// It only matches the query as a whole, but the position of the query in a fact doesn't matter.
// It searches all chats a user is part of.
func (store SQLStore) FindFacts(userID int, query string) (facts []Fact, err error) {
	rows, err := store.db.Query(findFactsStmt, userID, query)
	if err != nil {
		return facts, errors.Wrapf(err, "failed to query facts for userID %d and query '%s'", userID, query)
	}
	defer func() {
		closeErr := rows.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	for rows.Next() {
		var f Fact
		err = rows.Scan(&f.ID, &f.ChatID, &f.Content, &f.ChatTitle)
		if err != nil {
			return facts, errors.Wrap(err, "failed to scan row")
		}
		facts = append(facts, f)
	}
	err = rows.Err()
	return facts, errors.Wrap(err, "failed while querying facts")
}

// Close the underlying database connection.
func (store *SQLStore) Close() error {
	return errors.Wrap(store.db.Close(), "failed to close database")
}
