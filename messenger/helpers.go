package messenger

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/jorinvo/slangbrain/brain"
	"github.com/jorinvo/slangbrain/fbot"
	"github.com/jorinvo/slangbrain/translate"
	"github.com/jorinvo/slangbrain/user"
)

// Everything that is not in the unicode character classes
// for letters or numeric values
// See: http://www.fileformat.info/info/unicode/category/index.htm
var specialChars = regexp.MustCompile(`[^\p{Ll}\p{Lm}\p{Lo}\p{Lu}\p{Nd}\p{Nl}\p{No}\p{Mn}]`)

// inside () or [] or || or {} or <>
var inParantheses = regexp.MustCompile(`\(.*?\)|\[.*?\]|\|.*?\||\{.*?\}|\<.*?\>`)

// Keeping it simple for now.
// handleLinks() will find the false positives.
var matchURL = regexp.MustCompile(`https?://\S+\.\S+`)

// Change to menu mode.
// Return values can be passed directly to b.send().
func (b Bot) messageStartMenu(u user.User) (int64, string, []fbot.Reply, error) {
	if err := b.store.SetMode(u.ID, brain.ModeMenu); err != nil {
		return u.ID, u.Msg.Error, u.Rpl.MenuMode, err
	}
	return u.ID, u.Msg.Menu, u.Rpl.MenuMode, nil
}

// Send both welcome messages after each other.
func (b Bot) messageWelcome(u user.User) {
	if err := b.store.Register(u.ID); err != nil {
		b.err.Printf("failed to register user %d: %v", u.ID, err)
	}
	b.send(u.ID, fmt.Sprintf(u.Msg.Welcome1, u.Name()), nil, nil)
	time.Sleep(b.messageDelay)
	b.send(u.ID, u.Msg.Welcome2, nil, nil)
	time.Sleep(b.messageDelay)
	b.send(u.ID, u.Msg.Welcome3, nil, nil)
	time.Sleep(b.messageDelay)
	b.send(u.ID, u.Msg.Welcome4, nil, b.store.SetMode(u.ID, brain.ModeAdd))
}

// Change to study mode and find correct message.
// Return values can be passed directly to b.send().
func (b Bot) startStudy(u user.User) (int64, string, []fbot.Reply, error) {
	study, err := b.store.GetStudy(u.ID)
	if err != nil {
		return u.ID, u.Msg.Error, u.Rpl.StudyMode, err
	}
	// No studies ready
	if study.Total == 0 {
		// Go to menu mode
		if err = b.store.SetMode(u.ID, brain.ModeMenu); err != nil {
			return u.ID, u.Msg.Error, u.Rpl.StudyMode, err
		}
		// There are no studies yet
		if study.Next == 0 {
			return u.ID, u.Msg.StudyEmpty, u.Rpl.StudyEmpty, nil
		}
		// Display time until next study is ready
		msg := fmt.Sprintf(u.Msg.StudyDone, formatDuration(u.Msg, study.Next))
		isSubscribed, err := b.store.IsSubscribed(u.ID)
		if err != nil {
			b.err.Println(err)
		}
		if isSubscribed || err != nil {
			return u.ID, msg, u.Rpl.MenuMode, nil
		}
		// Ask to subscribe to notifications
		return u.ID, msg + "\n\n" + u.Msg.AskToSubscribe, u.Rpl.Subscribe, nil
	}
	// Send study to user
	return u.ID, fmt.Sprintf(u.Msg.StudyQuestion, study.Total, study.Explanation), u.Rpl.Show, nil
}

// Score current study and continue with next one.
// Return values can be passed directly to b.send().
func (b Bot) scoreAndStudy(u user.User, score int) (int64, string, []fbot.Reply, error) {
	err := b.store.ScoreStudy(u.ID, score)
	if err != nil {
		return u.ID, u.Msg.Error, u.Rpl.StudyMode, err
	}
	return b.startStudy(u)
}

// Send replies and log errors.
func (b Bot) send(id int64, reply string, buttons []fbot.Reply, err error) {
	if err != nil {
		b.err.Println(err)
	}
	if err = b.client.Send(id, reply, buttons); err != nil {
		b.err.Println("failed to send message:", err)
	}
}

// Format like "X hour[s], X minute[s]".
// Returns empty string for negativ durations.
func formatDuration(msg translate.Msg, d time.Duration) string {
	// Precision in minutes
	d = time.Duration(math.Ceil(float64(d)/float64(time.Minute))) * time.Minute
	s := ""
	h := d / time.Hour
	m := (d - h*time.Hour) / time.Minute
	if h > 1 {
		s += fmt.Sprintf("%d %s, ", h, msg.Hours)
	} else if h == 1 {
		s += msg.AnHour + ", "
	}
	if m > 1 {
		s += fmt.Sprintf("%d %s", m, msg.Minutes)
	} else if m > 0 {
		s += msg.AMinute
	} else if s != "" {
		// No minutes, only hours, remove trailing comma and space
		s = s[:len(s)-2]
	}
	return s
}

// Normalize two forms so user can choose to add parts in paranthesis or not.
// Case, space and punctuation are ignored.
func normPhrases(s string) (string, string) {
	return normPhrase(inParantheses.ReplaceAllString(s, "")), normPhrase(s)
}
func normPhrase(s string) string {
	return specialChars.ReplaceAllString(strings.ToLower(strings.TrimSpace(s)), "")
}

// Finds all links in a given text and returns them.
// Returns nil if no links could be found.
func getLinks(s string) []string {
	return matchURL.FindAllString(s, -1)
}
