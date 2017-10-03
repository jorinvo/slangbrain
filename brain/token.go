package brain

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	bolt "github.com/coreos/bbolt"
	"github.com/jorinvo/slangbrain/brain/bucket"
)

// GenerateToken creates and returns the token for a user.
// It can be used to authenticate the user after.
// If a token already exists, it is reused instead of generating a new one.
// There is no way to expire tokens for now, because links should be shareable and token should be useable for automation.
func (store Store) GenerateToken(id int64) (string, error) {
	var token string
	err := store.db.Update(func(tx *bolt.Tx) error {
		bid := itob(id)
		bu := tx.Bucket(bucket.AuthUsers)
		bt := tx.Bucket(bucket.AuthTokens)
		now := time.Now()

		// Lookup existing
		if v := bu.Get(bid); v != nil {
			token = string(v[8:])
			return nil
		}

		// Or create new
		t, err := random(authTokenLength)
		if err != nil {
			return fmt.Errorf("failed to generate token: %v", err)
		}
		token = t
		if err := bt.Put([]byte(token), bid); err != nil {
			return err
		}

		return bu.Put(bid, append(itob(now.Unix()), []byte(token)...))
	})

	if err != nil {
		return "", fmt.Errorf("failed to get auth token for %d: %v", id, err)
	}
	return token, nil
}

// LookupToken returns the chat id a token is registered for.
// Returns ErrNotFound if token is invalid.
func (store Store) LookupToken(token string) (int64, error) {
	var id int64
	err := store.db.View(func(tx *bolt.Tx) error {
		i := tx.Bucket(bucket.AuthTokens).Get([]byte(token))
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
		return id, fmt.Errorf("failed to lookup auth token for %d: %v", id, err)
	}
	return id, nil
}

func random(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return base64.URLEncoding.EncodeToString(b), err
}
