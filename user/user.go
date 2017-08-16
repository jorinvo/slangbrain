package user

import (
	"fmt"
	"log"
	"time"

	"github.com/jorinvo/slangbrain/brain"
	"github.com/jorinvo/slangbrain/common"
	"github.com/jorinvo/slangbrain/translate"
)

// User stores relevent information for the current request.
// It is used for handling messages and payloads.
// A profile and content in the correct language are loaded.
type User struct {
	ID int64
	common.Profile
	translate.Content
}

type fetcher func(int64) (common.Profile, error)

// Get a user with profile and content loaded.
// Logs errors.
// Check fields of profile before using them.
// It is not guaranteed that they are successfully loaded.
func Get(id int64, s brain.Store, l *log.Logger, t translate.Translator, f fetcher) User {
	p, err := getProfile(id, s, l, f)
	if err != nil {
		l.Printf("failed to get profile for id %d: %v", id, err)
	}
	return User{
		ID:      id,
		Profile: p,
		Content: t.Load(p.Locale()),
	}
}

// Get profile from cache or fetch and update cache.
func getProfile(id int64, s brain.Store, l *log.Logger, f fetcher) (common.Profile, error) {
	// Try cache
	p, err := s.GetProfile(id)
	if err == nil {
		return p, nil
	}
	if err != brain.ErrNotFound {
		l.Println(err)
	}
	if f == nil {
		return p, err
	}
	// Fetch from Facebook
	p, err = f(id)
	if err != nil {
		return p, fmt.Errorf("failed to fetch profile: %v", err)
	}
	// Update cache
	if err := s.SetProfile(id, p, time.Now()); err != nil {
		l.Println(err)
	}
	return p, nil
}
