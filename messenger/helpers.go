package messenger

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jorinvo/slangbrain/brain"
	"github.com/jorinvo/slangbrain/common"
	"github.com/jorinvo/slangbrain/fbot"
)

// Change to menu mode.
// Return values can be passed directly to b.send().
func (b Bot) messageStartMenu(u user) (int64, string, []fbot.Button, error) {
	if err := b.store.SetMode(u.ID, brain.ModeMenu); err != nil {
		return u.ID, u.Msg.Error, u.Btn.MenuMode, err
	}
	return u.ID, u.Msg.Menu, u.Btn.MenuMode, nil
}

// Send both welcome messages after each other.
func (b Bot) messageWelcome(u user) {
	if err := b.store.Register(u.ID); err != nil {
		b.err.Printf("failed to register user %d: %v", u.ID, err)
	}
	b.send(u.ID, fmt.Sprintf(u.Msg.Welcome1, u.Name()), nil, nil)
	time.Sleep(b.welcomeWait)
	b.send(u.ID, u.Msg.Welcome2, nil, b.store.SetMode(u.ID, brain.ModeAdd))
}

// Change to study mode and find correct message.
// Return values can be passed directly to b.send().
func (b Bot) startStudy(u user) (int64, string, []fbot.Button, error) {
	study, err := b.store.GetStudy(u.ID)
	if err != nil {
		return u.ID, u.Msg.Error, u.Btn.StudyMode, err
	}
	// No studies ready
	if study.Total == 0 {
		// Go to menu mode
		if err = b.store.SetMode(u.ID, brain.ModeMenu); err != nil {
			return u.ID, u.Msg.Error, u.Btn.StudyMode, err
		}
		// There are no studies yet
		if study.Next == 0 {
			return u.ID, u.Msg.StudyEmpty, u.Btn.StudyEmpty, nil
		}
		// Display time until next study is ready
		msg := fmt.Sprintf(u.Msg.StudyDone, formatDuration(study.Next))
		isSubscribed, err := b.store.IsSubscribed(u.ID)
		if err != nil {
			b.err.Println(err)
		}
		if isSubscribed || err != nil {
			return u.ID, msg, u.Btn.MenuMode, nil
		}
		// Ask to subscribe to notifications
		return u.ID, msg + u.Msg.AskToSubscribe, u.Btn.Subscribe, nil
	}
	// Send study to user
	return u.ID, fmt.Sprintf(u.Msg.StudyQuestion, study.Total, study.Explanation), u.Btn.Show, nil
}

// Score current study and continue with next one.
// Return values can be passed directly to b.send().
func (b Bot) scoreAndStudy(u user, score int) (int64, string, []fbot.Button, error) {
	err := b.store.ScoreStudy(u.ID, score)
	if err != nil {
		return u.ID, u.Msg.Error, u.Btn.StudyMode, err
	}
	return b.startStudy(u)
}

// Send replies and log errors.
func (b Bot) send(id int64, reply string, buttons []fbot.Button, err error) {
	if err != nil {
		b.err.Println(err)
	}
	if err = b.client.Send(id, reply, buttons); err != nil {
		b.err.Println("failed to send message:", err)
	}
}

// Get a user with profile and content loaded.
func (b Bot) getUser(id int64) user {
	p, err := b.getProfile(id)
	if err != nil {
		b.err.Printf("failed to get profile for id %d: %v", id, err)
	}
	return user{
		ID:      id,
		Profile: p,
		Content: b.content.Load(p.Locale()),
	}
}

// Get profile from cache or fetch and update cache.
func (b Bot) getProfile(id int64) (common.Profile, error) {
	// Try cache
	p, err := b.store.GetProfile(id)
	if err == nil {
		return p, nil
	}
	if err != brain.ErrNotFound {
		b.err.Println(err)
	}
	// Fetch from Facebook
	p, err = b.client.GetProfile(id)
	if err != nil {
		return p, fmt.Errorf("failed to fetch profile: %v", err)
	}
	// Update cache
	if err := b.store.SetProfile(id, p); err != nil {
		b.err.Println(err)
	}
	return p, nil
}

// Format like "X hour[s] X minute[s]".
// Returns empty string for negativ durations.
func formatDuration(d time.Duration) string {
	// Precision in minutes
	d = time.Duration(math.Ceil(float64(d)/float64(time.Minute))) * time.Minute
	s := ""
	h := d / time.Hour
	m := (d - h*time.Hour) / time.Minute
	if h > 1 {
		s += fmt.Sprintf("%d", h) + " hours "
	} else if h == 1 {
		s += "1 hour "
	}
	if m > 1 {
		s += fmt.Sprintf("%d", m) + " minutes"
	} else if m > 0 {
		s += "1 minute"
	} else if s != "" {
		// No minutes, only hours, remove trailing space
		s = s[:len(s)-1]
	}
	return s
}

// Normalize two forms so user can choose add parts in paranthesis or not.
// Case, space and punctuation are ignored.
func normPhrases(s string) (string, string) {
	return normPhrase(inParantheses.ReplaceAllString(s, "")), normPhrase(s)
}
func normPhrase(s string) string {
	return specialChars.ReplaceAllString(strings.ToLower(strings.TrimSpace(s)), "")
}
