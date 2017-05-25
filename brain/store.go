package brain

import (
	"database/sql"
	"fmt"
	"io/ioutil"
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
		return store, fmt.Errorf("failed to initialize database adapter: %v", err)
	}

	if err = db.Ping(); err != nil {
		return store, fmt.Errorf("failed to ping database: %v", err)
	}

	setupStmt, err := ioutil.ReadFile(setupSQL)
	if err != nil {
		return store, fmt.Errorf("failed to read setup script from %s: %v", setupSQL, err)
	}
	_, err = db.Exec(string(setupStmt))
	if err != nil {
		return store, fmt.Errorf("failed to setup tables: %v", err)
	}

	return store, nil
}

// AddPhrase stores a new phrase.
func (store Store) AddPhrase(chatID int64, phrase, explanation string) error {
	var phraseID int
	row := store.db.QueryRow(addPhraseStmt, chatID, phrase, explanation)
	if err := row.Scan(&phraseID); err != nil {
		return fmt.Errorf("failed to add phrase for chatID %d: %s - %s: %v", chatID, phrase, explanation, err)
	}
	for _, mode := range Studymodes {
		_, err := store.db.Exec(addStudyStmt, phraseID, mode)
		if err != nil {
			return fmt.Errorf("failed to add phrase for chatID %d: %s - %s: %v", chatID, phrase, explanation, err)
		}
	}
	return nil
}

// GetMode fetches the mode for a chat.
func (store Store) GetMode(chatID int64) (Mode, error) {
	var m Mode
	row := store.db.QueryRow(getModeStmt, chatID)
	if err := row.Scan(&m); err != nil {
		return m, fmt.Errorf("failed to get mode for chatID %d: %v", chatID, err)
	}
	return m, nil
}

// SetMode updates the mode for a chat.
func (store Store) SetMode(chatID int64, mode Mode) error {
	_, err := store.db.Exec(setModeStmt, chatID, mode)
	if err != nil {
		return fmt.Errorf("failed to set mode for chatID %d: %d: %v", chatID, mode, err)
	}
	return nil
}

// GetStudy ...
func (store Store) GetStudy(chatID int64) (Study, error) {
	var s Study
	row := store.db.QueryRow(getStudyStmt, chatID)
	err := row.Scan(&s.ID, &s.Mode, &s.Phrase, &s.Explanation, &s.Total)
	if err == sql.ErrNoRows {
		err = nil
	}
	if err != nil {
		return s, fmt.Errorf("failed to study with chatID %d: %v", chatID, err)
	}
	return s, nil
}

// ScoreStudy ...
func (store Store) ScoreStudy(chatID int64, score Score) error {
	_, err := store.db.Exec(scoreStmt, chatID, score)
	if err != nil {
		return fmt.Errorf("failed to score study for chatID %d: %d: %v", chatID, score, err)
	}
	return nil
}

// CountStudies ...
func (store Store) CountStudies(chatID int64) (int, error) {
	count := 0
	row := store.db.QueryRow(countStudiesStmt, chatID)
	err := row.Scan(&count)
	if err == sql.ErrNoRows {
		err = nil
	}
	if err != nil {
		return count, fmt.Errorf("failed to count studies for chatID %d: %v", chatID, err)
	}
	return count, nil
}

// Close the underlying database connection.
func (store *Store) Close() error {
	if err := store.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %v", err)
	}
	return nil
}
