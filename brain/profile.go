package brain

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/jorinvo/slangbrain/common"
)

// Wrap data as common.Profile.
type profile struct {
	data profileData
}

type profileData struct {
	Name     string
	Locale   string
	Timezone float64
}

func (p profile) Name() string      { return p.data.Name }
func (p profile) Locale() string    { return p.data.Locale }
func (p profile) Timezone() float64 { return p.data.Timezone }

// GetProfile fetches a cached profile.
// Returns ErrNotFound if none found.
func (store Store) GetProfile(chatID int64) (common.Profile, error) {
	var p profileData
	err := store.db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(bucketProfiles).Get(itob(chatID))
		if v == nil {
			return ErrNotFound
		}
		return gob.NewDecoder(bytes.NewReader(v)).Decode(&p)
	})
	if err != nil {
		if err == ErrNotFound {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get profile for chatID %d: %v", chatID, err)
	}
	return profile{p}, nil
}

// SetProfile caches a profile.
func (store Store) SetProfile(chatID int64, p common.Profile) error {
	data := profileData{
		Name:     p.Name(),
		Locale:   p.Locale(),
		Timezone: p.Timezone(),
	}
	err := store.db.Update(func(tx *bolt.Tx) error {
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(data); err != nil {
			return err
		}
		return tx.Bucket(bucketProfiles).Put(itob(chatID), buf.Bytes())
	})
	if err != nil {
		return fmt.Errorf("failed to set profile for chatID %d: %v: %v", chatID, data, err)
	}
	return nil
}
