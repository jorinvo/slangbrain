package brain

import (
	"encoding/json"
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/jorinvo/slangbrain/common"
)

// Wrap data as common.Profile.
type profile struct {
	data profileData
}

// Use short JSON names to save disc space.
type profileData struct {
	Name     string  `json:"n,omit_empty"`
	Locale   string  `json:"l,omit_empty"`
	Timezone float64 `json:"t,omit_empty"`
}

func (p profile) Name() string      { return p.data.Name }
func (p profile) Locale() string    { return p.data.Locale }
func (p profile) Timezone() float64 { return p.data.Timezone }
func (p *profile) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &p.data)
}

// GetProfile fetches a cached profile.
// Returns ErrNotFound if none found.
func (store Store) GetProfile(chatID int64) (common.Profile, error) {
	var p profile
	err := store.db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(bucketProfiles).Get(itob(chatID))
		if v == nil {
			return ErrNotFound
		}
		return json.Unmarshal(v, &p)
	})
	if err != nil {
		if err == ErrNotFound {
			return p, err
		}
		return p, fmt.Errorf("failed to get profile for chatID %d: %v", chatID, err)
	}
	return p, nil
}

// SetProfile caches a profile.
func (store Store) SetProfile(chatID int64, p common.Profile) error {
	data := profileData{
		Name:     p.Name(),
		Locale:   p.Locale(),
		Timezone: p.Timezone(),
	}
	err := store.db.Update(func(tx *bolt.Tx) error {
		buf, err := json.Marshal(data)
		if err != nil {
			return err
		}
		return tx.Bucket(bucketProfiles).Put(itob(chatID), buf)
	})
	if err != nil {
		return fmt.Errorf("failed to set profile for chatID %d: %v: %v", chatID, data, err)
	}
	return nil
}
