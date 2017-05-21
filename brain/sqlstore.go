package brain

import (
	"database/sql"

	"github.com/pkg/errors"
)

var (
	setupStmt = `
		CREATE TABLE IF NOT EXISTS phrases (
			id INTEGER PRIMARY KEY,
			chatid BIGINT,
			'foreign' TEXT,
			mother TEXT
		);
	`

	addPhraseStmt = `
		INSERT INTO phrases (chatid, 'foreign', mother) VALUES ($1, $2, $3)
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

// AddPhrase stores a new phrase.
func (store SQLStore) AddPhrase(chatID int64, foreign, mother string) error {
	_, err := store.db.Exec(addPhraseStmt, chatID, foreign, mother)
	return errors.Wrapf(err, "failed to add phrase for chatID %d: %s - %s", chatID, foreign, mother)
}

// Close the underlying database connection.
func (store *SQLStore) Close() error {
	return errors.Wrap(store.db.Close(), "failed to close database")
}
