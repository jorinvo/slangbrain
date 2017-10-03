package bot

import (
	"fmt"
	"time"

	"github.com/jorinvo/slangbrain/brain"
	"github.com/jorinvo/slangbrain/scope"
)

// Start a timer to notify the given chat.
// Only works when chat has notifications enabled
// and has added some phrases already.
func (b Bot) scheduleNotify(id int64) {
	if b.notifyTimers == nil {
		return
	}

	isSubscribed, err := b.store.IsSubscribed(id)
	if err != nil {
		b.err.Println(err)
		return
	}
	if !isSubscribed {
		return
	}

	if timer := b.notifyTimers[id]; timer != nil {
		// Don't care if timer is active or not
		_ = timer.Stop()
	}
	u := scope.Get(id, b.store, b.err, b.content, b.client.GetProfile)
	d, count, err := b.store.GetNotifyTime(id, u.Timezone())
	if err != nil {
		b.err.Println(err)
		return
	}
	if count <= 1 {
		return
	}

	b.info.Printf("Notify %d in %s with %d due studies", id, d.String(), count)
	b.notifyTimers[id] = time.AfterFunc(d, func() {
		b.notify(id, count)
	})
}

func (b Bot) notify(id int64, count int) {
	u := scope.Get(id, b.store, b.err, b.content, b.client.GetProfile)
	msg := fmt.Sprintf(u.Msg.StudyNotification, u.Name(), count)
	if err := b.store.SetMode(id, brain.ModeMenu); err != nil {
		b.err.Printf("failed to activate menu mode while notifying %d: %v", u.ID, err)
	}
	if err := b.client.Send(id, msg, u.Rpl.StudiesDue); err != nil {
		b.err.Printf("failed to notify user %d: %v", u.ID, err)
	}
	b.info.Printf("Notified %s (%d) with %d due studies", u.Name(), u.ID, count)
	// Track last sending of a notification
	// to stop sending notifications
	// when user hasn't read the last notification.
	if err := b.store.TrackNotify(u.ID, time.Now()); err != nil {
		b.err.Println(err)
	}
}
