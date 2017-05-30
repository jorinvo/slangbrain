package messenger

import "github.com/jorinvo/messenger"

var (
	buttonStudyDone = button("done studying", payloadStartMenu)
	// Teacher emoji
	buttonStudy = button("\U0001F468\u200D\U0001F3EB study", payloadStartStudy)
	// Plus sign emoji
	buttonAdd = button("\u2795 phrases", payloadStartAdd)
	// Waving hand emoji
	buttonDone = button("\u2714 done", payloadIdle)
)

var (
	buttonsMenuMode = []messenger.QuickReply{
		buttonStudy,
		buttonAdd,
		button("\u2753 help", payloadShowHelp),
		buttonDone,
	}
	buttonsHelp = []messenger.QuickReply{
		buttonStudy,
		buttonAdd,
		buttonDone,
	}
	buttonsAddMode = []messenger.QuickReply{
		button("stop adding", payloadStartMenu),
	}
	buttonsStudyMode = []messenger.QuickReply{
		buttonStudyDone,
	}
	buttonsShow = []messenger.QuickReply{
		buttonStudyDone,
		button("\U0001F449 show phrase", payloadShowStudy),
	}
	buttonsScore = []messenger.QuickReply{
		// Ok thumb down emoji
		button("\U0001F44E didn't know", payloadScoreBad),
		button("soso", payloadScoreOk),
		// Ok hand emoji
		button("\U0001F44C easy", payloadScoreGood),
	}
	buttonsStudiesDue = []messenger.QuickReply{
		buttonStudy,
		button("not now", payloadStartMenu),
	}
)

func button(title, payload string) messenger.QuickReply {
	return messenger.QuickReply{
		ContentType: "text",
		Title:       title,
		Payload:     payload,
	}
}
