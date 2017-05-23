package brain

import (
	"database/sql"
	"io/ioutil"

	"github.com/pkg/errors"
)

const setupSQL = "setup.sql"

// Store implements a Store as classical SQL database.
// It uses the standard database/sql package and supports different backends.
// Current implementation is tested with SQLite.
type Store struct {
	db *sql.DB
}

// CreateStore returns a new SQLStore with a database already setup.
// It forwards the arguments to the standard database/sql package.
func CreateStore(driverName, dataSourceName string) (Store, error) {
	db, err := sql.Open(driverName, dataSourceName)
	store := Store{db}
	if err != nil {
		return store, errors.Wrap(err, "failed to initialize database adapter")
	}

	setupStmt, err := ioutil.ReadFile(setupSQL)
	if err != nil {
		return store, errors.Wrapf(err, "failed to read setup script from %s", setupSQL)
	}
	_, err = db.Exec(string(setupStmt))
	if err != nil {
		return store, errors.Wrap(err, "failed to setup tables")
	}

	err = db.Ping()
	return store, errors.Wrap(err, "failed to connect to database")
}

// AddPhrase stores a new phrase.
func (store Store) AddPhrase(chatID int64, phrase, explanation string) error {
	r, err := store.db.Exec(addPhraseStmt, chatID, phrase, explanation)
	if err != nil {
		return errors.Wrapf(err, "failed to add phrase for chatID %d: %s - %s", chatID, phrase, explanation)
	}
	phraseID, err := r.LastInsertId()
	if err != nil {
		return errors.Wrapf(err, "failed to add phrase for chatID %d: %s - %s", chatID, phrase, explanation)
	}

	for mode := range Studymodes {
		_, err = store.db.Exec(addStudyStmt, phraseID, mode)
		if err != nil {
			return errors.Wrapf(err, "failed to add phrase for chatID %d: %s - %s", chatID, phrase, explanation)
		}
	}
	return errors.Wrapf(err, "failed to add phrase for chatID %d: %s - %s", chatID, phrase, explanation)
}

// GetMode fetches the mode for a chat.
func (store Store) GetMode(chatID int64) (Mode, error) {
	var m Mode
	row := store.db.QueryRow(getModeStmt, chatID)
	err := row.Scan(&m)
	return m, errors.Wrapf(err, "failed to get mode for chatID %d", chatID)
}

// SetMode updates the mode for a chat.
func (store Store) SetMode(chatID int64, mode Mode) error {
	_, err := store.db.Exec(setModeStmt, chatID, mode)
	return errors.Wrapf(err, "failed to set mode for chatID %d: %d", chatID, mode)
}

// GetStudy ...
func (store Store) GetStudy(chatID int64) (Study, error) {
	var s Study
	row := store.db.QueryRow(getStudyStmt, chatID)
	err := row.Scan(&s.ID, &s.Mode, &s.Phrase, &s.Explanation, &s.Total)
	if err == sql.ErrNoRows {
		err = nil
	}
	return s, errors.Wrapf(err, "failed to study with chatID %d", chatID)
}

// ScoreStudy ...
func (store Store) ScoreStudy(chatID int64, score Score) error {
	_, err := store.db.Exec(scoreStmt, chatID, score)
	return errors.Wrapf(err, "failed to score study for chatID %d: %d", chatID, score)
}

// Close the underlying database connection.
func (store *Store) Close() error {
	return errors.Wrap(store.db.Close(), "failed to close database")
}
