package brain

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/boltdb/bolt"
)

// GenerateToken creates and returns the token for a user.
// It can be used to authenticate the user after.
func (store Store) GenerateToken(id int64) (string, error) {
	token, err := random(authTokenLength)
	if err != nil {
		return "", fmt.Errorf("failed to generate manage token for %d: %v", id, err)
	}
	err = store.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketAuthTokens).Put([]byte(token), itob(id))
	})
	if err != nil {
		return "", fmt.Errorf("failed to write manage token for %d: %v", id, err)
	}
	return token, nil
}

// LookupToken returns the chat id a token is registered for.
// Returns ErrNotFound if token is invalid.
func (store Store) LookupToken(token string) (int64, error) {
	var id int64
	err := store.db.View(func(tx *bolt.Tx) error {
		i := tx.Bucket(bucketAuthTokens).Get([]byte(token))
		if i == nil {
			return ErrNotFound
		}
		id = btoi(i)
		return nil
	})
	if err != nil {
		if err == ErrNotFound {
			return id, err
		}
		return id, fmt.Errorf("failed to write manage token for %d: %v", id, err)
	}
	return id, nil
}

func random(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return base64.URLEncoding.EncodeToString(b), err
}
