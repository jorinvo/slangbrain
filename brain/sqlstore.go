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
			chatid BIGINT UNIQUE,
			title TEXT,
			mode INTEGER DEFAULT 0
		);
		CREATE TABLE IF NOT EXISTS chatmembers (
			id INTEGER PRIMARY KEY,
			userid INTEGER,
			chatid REFERENCES chats
		);
		CREATE UNIQUE INDEX IF NOT EXISTS uniqidcombi ON chatmembers(chatid, userid);
	`

	saveChatmemberStmt = `
		REPLACE INTO chatmembers (chatid, userid) VALUES ($1, $2)
	`

	updateChatStmt = `
		REPLACE INTO chats (chatid, title, mode) VALUES ($1, $2, $3)
	`

	selectModeStmt = `
		SELECT mode FROM chats WHERE chatid = $1
	`

	setModeStmt = `
		UPDATE chats SET mode = $2 WHERE chatid = $1
	`

	addFactStmt = `
		INSERT INTO facts (chatid, content) VALUES ($1, $2)
	`

	findFactsStmt = `
		SELECT facts.content, facts.chatid, chats.title
		FROM facts
		JOIN chats ON facts.chatid = chats.chatid
		JOIN chatmembers ON facts.chatid = chatmembers.chatid
		WHERE chatmembers.userid = $1
		AND facts.content LIKE '%' || $2 || '%'
	`
)

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

// UseChat updates and returns a chat.
// It saves a relation between a userID and a chatID.
// It updates the chat title.
// Relations are stored unique - even after calling SaveChat multiple times.
func (store SQLStore) UseChat(chatID int64, userID int, chatTitle string) (Chat, error) {
	chat := Chat{ID: chatID, Title: chatTitle}

	// Get mode first to not overwrite it later
	row := store.db.QueryRow(selectModeStmt, chatID)
	err := row.Scan(&chat.Mode)
	if err != nil {
		return chat, errors.Wrapf(err, "failed to get mode for chatID %d", chatID)
	}

	_, err = store.db.Exec(updateChatStmt, chat.ID, chat.Title, chat.Mode)
	if err != nil {
		return chat, errors.Wrapf(err, "failed to update chat %d", chatID)
	}

	_, err = store.db.Exec(saveChatmemberStmt, chatID, userID)
	return chat, errors.Wrapf(err, "failed to save chat member for chatID %d and userID %d", chatID, userID)
}

// SetMode updates the mode of a chat
func (store SQLStore) SetMode(chatID int64, mode Mode) error {
	_, err := store.db.Exec(setModeStmt, chatID, mode)
	return errors.Wrapf(err, "failed to set mode for chatID %d", chatID)
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
		f := Fact{}
		err = rows.Scan(&f.Content, &f.Chat.ID, &f.Chat.Title)
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
