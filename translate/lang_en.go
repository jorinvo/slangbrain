package translate

// en provides content in English.
func en() (Msg, buttonLabels) {
	m := Msg{
		// Greeting is currently limited to 160 chars.
		// {{user_first_name}} is replaced by Facebook.
		Greeting: `Hi {{user_first_name}}! Slangbrain helps you with our language studies.
Master the language you encounter daily instead of limiting yourself to a textbook.`,
		Menu: `What would you like to do next?
Please use the buttons below.`,
		Help: "How can I help you?",
		Idle: "Good, just send me a " + iconThumbsup + " to continue with your studies.",
		Add: `Please send me a phrase and its explanation.
Separate them with a linebreak.`,
		Welcome1: `Hello %s!

Whenever you pick up a new phrase, just add it to your Slangbrain and remember it forever.

You begin by adding phrases and later Slangbrain will test your memories in a natural schedule.`,
		Welcome2: `Please send me a phrase and its explanation.
Separate them with a linebreak.
Don't worry if you send something wrong. You can delete phrases later.

If your mother tongue is English and you're studying Spanish, a message would look like this:

Hola
Hello

Give it a try:`,
		Error:              "Sorry, something went wrong.",
		ExplanationMissing: "The phrase is missing an explanation. Please send it again with explanation.",
		PhraseMissing:      "Please send a phrase.",
		StudyDone: `Congrats, you finished all your studies for now!
Come back in %s.`,
		StudyCorrect: "Correct!",
		StudyWrong: `Sorry, the right version is:

%s`,
		StudyEmpty: `You have added no phrases yet.
Click the button below and get started.`,
		StudyQuestion: `%d. Do you remember how to say this?

%s

Use the buttons or type the phrase.`,
		ExplanationExists: `You already saved a phrase with the same explanation:
%s
%s

Please send it again with an explanation you can distinguish from the existing one.`,
		AddDone: `Saved phrase:
%s

With explanation:
%s`,
		AddNext:           "Add next phrase.",
		StudyNotification: `Hey %s, you have %d phrases ready for review!`,
		ConfirmDelete:     "Are you sure, you want to delete this phrase?",
		Deleted:           "The phrase has been deleted. Let's continue studying other phrases.",
		CancelDelete:      "Good, let's keep that phrase and continue studying.",
		AskToSubscribe: `

Would you like me to send you a message when there are phrases ready for studying?`,
		Subscribed: `Good, I will send you a message when your phrases are ready.

What would you like to do next?
Please use the buttons below.`,
		ConfirmUnsubscribe: `Sure, you won't receive any more notifications.

What would you like to do next?
Please use the buttons below.`,
		Unsubscribed: `Sure, you won't receive any notifications.

What would you like to do next?
Please use the buttons below.`,
		Fedback:      "If you run into a problem, have any feedback for the people behind Slangbrain or just like to say hello, you can send a message now and we will get back to you as soon as possible.",
		FeedbackDone: "Thanks %s, you will hear from us soon.",
	}

	b := buttonLabels{
		StudyDone:         "done studying",
		Study:             "study",
		Add:               "phrases",
		Done:              "done",
		Help:              "help",
		SubscribeConfirm:  "sounds good",
		SubscribeDeny:     "no thanks",
		StopNotifications: "stop notifications",
		SendFeedback:      "send feedback",
		QuitHelp:          "all good",
		CancelFeedback:    "cancel",
		StopAdding:        "stop adding",
		ShowPhrase:        "show phrase",
		ScoreBad:          "didn't know",
		ScoreGood:         "got it",
		StudyNotNow:       "not now",
		ConfirmDelete:     "delete phrase",
		CancelDelete:      "cancel",
	}

	return m, b
}
