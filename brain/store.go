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

	setupStmt, err := ioutil.ReadFile(setupSQL)
	if err != nil {
		return store, fmt.Errorf("failed to read setup script from %s: %v", setupSQL, err)
	}
	_, err = db.Exec(string(setupStmt))
	if err != nil {
		return store, fmt.Errorf("failed to setup tables: %v", err)
	}

	err = db.Ping()
	return store, fmt.Errorf("failed to connect to database: %v", err)
}

// AddPhrase stores a new phrase.
func (store Store) AddPhrase(chatID int64, phrase, explanation string) error {
	r, err := store.db.Exec(addPhraseStmt, chatID, phrase, explanation)
	if err != nil {
		return fmt.Errorf("failed to add phrase for chatID %d: %s - %s: %v", chatID, phrase, explanation, err)
	}
	phraseID, err := r.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to add phrase for chatID %d: %s - %s: %v", chatID, phrase, explanation, err)
	}

	for _, mode := range Studymodes {
		_, err = store.db.Exec(addStudyStmt, phraseID, mode)
		if err != nil {
			return fmt.Errorf("failed to add phrase for chatID %d: %s - %s: %v", chatID, phrase, explanation, err)
		}
	}
	return fmt.Errorf("failed to add phrase for chatID %d: %s - %s: %v", chatID, phrase, explanation, err)
}

// GetMode fetches the mode for a chat.
func (store Store) GetMode(chatID int64) (Mode, error) {
	var m Mode
	row := store.db.QueryRow(getModeStmt, chatID)
	err := row.Scan(&m)
	return m, fmt.Errorf("failed to get mode for chatID %d: %v", chatID, err)
}

// SetMode updates the mode for a chat.
func (store Store) SetMode(chatID int64, mode Mode) error {
	_, err := store.db.Exec(setModeStmt, chatID, mode)
	return fmt.Errorf("failed to set mode for chatID %d: %d: %v", chatID, mode, err)
}

// GetStudy ...
func (store Store) GetStudy(chatID int64) (Study, error) {
	var s Study
	row := store.db.QueryRow(getStudyStmt, chatID)
	err := row.Scan(&s.ID, &s.Mode, &s.Phrase, &s.Explanation, &s.Total)
	if err == sql.ErrNoRows {
		err = nil
	}
	return s, fmt.Errorf("failed to study with chatID %d: %v", chatID, err)
}

// ScoreStudy ...
func (store Store) ScoreStudy(chatID int64, score Score) error {
	_, err := store.db.Exec(scoreStmt, chatID, score)
	return fmt.Errorf("failed to score study for chatID %d: %d: %v", chatID, score, err)
}

// CountStudies ...
func (store Store) CountStudies(chatID int64) (int, error) {
	count := 0
	row := store.db.QueryRow(countStudiesStmt, chatID)
	err := row.Scan(&count)
	if err == sql.ErrNoRows {
		err = nil
	}
	return count, fmt.Errorf("failed to count studies for chatID %d: %v", chatID, err)
}

// Close the underlying database connection.
func (store *Store) Close() error {
	return fmt.Errorf("failed to close database: %v", store.db.Close())
}
