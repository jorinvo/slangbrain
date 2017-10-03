package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/jorinvo/slangbrain/brain"
	"github.com/jorinvo/slangbrain/scope"
)

func (b Bot) handleMessage(u scope.User, msg string) {
	// If message contains links, handle them instead of whatever would be next
	if links := getLinks(msg); links != nil {
		b.handleLinks(u, links)
		return
	}

	mode, err := b.store.GetMode(u.ID)
	if err != nil {
		b.send(u.ID, u.Msg.Error, u.Rpl.MenuMode, fmt.Errorf("failed to get mode for id %v: %v", u.ID, err))
		return
	}
	switch mode {
	case brain.ModeStudy:
		study, err := b.store.GetStudy(u.ID)
		if err != nil {
			b.send(u.ID, u.Msg.Error, u.Rpl.StudyMode, fmt.Errorf("failed to get study: %v", err))
			return
		}
		// Score user unput and pick appropriate reply
		msgNormalizedA, msgNormalizedB := normPhrases(msg)
		if msgNormalizedA == "" {
			study, err := b.store.GetStudy(u.ID)
			if err != nil {
				b.send(u.ID, u.Msg.Error, u.Rpl.Show, fmt.Errorf("failed to get study: %v", err))
				return
			}
			b.send(u.ID, study.Phrase, u.Rpl.Score, nil)
			return
		}
		var score = 1
		reply := u.Msg.StudyCorrect
		phraseNormalizedA, phraseNormalizedB := normPhrases(study.Phrase)
		if msgNormalizedA != phraseNormalizedA && msgNormalizedB != phraseNormalizedB {
			score = -2
			reply = fmt.Sprintf(u.Msg.StudyWrong, study.Phrase)
		}
		b.send(u.ID, reply, nil, nil)
		b.send(b.scoreAndStudy(u, score))

	case brain.ModeAdd:
		parts := strings.SplitN(strings.TrimSpace(msg), "\n", 2)
		phrase := strings.TrimSpace(parts[0])
		if phrase == "" {
			b.send(u.ID, u.Msg.PhraseMissing, u.Rpl.AddMode, nil)
			return
		}
		if len(parts) == 1 {
			b.send(u.ID, u.Msg.ExplanationMissing, u.Rpl.AddMode, nil)
			return
		}
		explanation := strings.TrimSpace(parts[1])

		// Check for existing explanation
		p, err := b.store.FindPhrase(u.ID, func(p brain.Phrase) bool {
			return p.Explanation == explanation
		})
		if err != nil {
			b.send(u.ID, u.Msg.Error, nil, fmt.Errorf("failed to lookup phrase: %v", err))
			return
		}
		if p.Phrase != "" {
			b.send(u.ID, fmt.Sprintf(u.Msg.ExplanationExists, p.Phrase, p.Explanation), u.Rpl.AddMode, nil)
			return
		}

		// Save phrase
		if err = b.store.AddPhrase(u.ID, phrase, explanation, time.Now()); err != nil {
			b.send(u.ID, u.Msg.Error, u.Rpl.AddMode, fmt.Errorf("failed to save phrase: %v", err))
			return
		}

		b.send(u.ID, fmt.Sprintf(u.Msg.AddDone, phrase, explanation), nil, nil)
		b.send(u.ID, u.Msg.AddNext, u.Rpl.AddMode, nil)

	case brain.ModeGetStarted:
		b.messageWelcome(u, "")

	case brain.ModeFeedback:
		b.feedback <- Feedback{ChatID: u.ID, Username: u.Name(), Message: msg}
		b.send(u.ID, fmt.Sprintf(u.Msg.FeedbackDone, u.Name()), nil, nil)
		b.send(b.messageStartMenu(u))

	default:
		b.feedback <- Feedback{
			ChatID:   u.ID,
			Username: u.Name(),
			Message:  msg,
			Channel:  slackUnhandled,
		}
		b.send(b.messageStartMenu(u))
	}
}
