package translate

// th provides content in Thai.
func th() (Msg, labels, Web) {
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

If your mother tongue is English and you're studying Spanish, a message would look like this:`,
		Welcome3: `Hola
Hello`,
		Welcome4:           "Now it's your turn:",
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
		AddNext:            "Add next phrase.",
		StudyNotification:  `Hey %s, you have %d phrases ready for review!`,
		ConfirmDelete:      "Are you sure, you want to delete this phrase?",
		Deleted:            "The phrase has been deleted. Let's continue studying other phrases.",
		CancelDelete:       "Good, let's keep that phrase and continue studying.",
		AskToSubscribe:     `Would you like me to send you a message when there are phrases ready for studying?`,
		Subscribed:         `Good, I will send you a message when your phrases are ready.`,
		ConfirmUnsubscribe: `Sure, you won't receive any more notifications.`,
		DenySubscribe:      `Sure, you won't receive any notifications.`,
		Feedback:           "If you run into a problem, have any feedback for the people behind Slangbrain or just like to say hello, you can send a message now and we will get back to you as soon as possible.",
		FeedbackDone:       "Thanks %s, you will hear from us soon.",
		ImportHelp1: `You can add many phrases at once by sending a CSV file to Slangbrain.
The file needs to end with '.csv' and it needs to have 2 columns, the first one is for  phrases, the second for their explanations.
Don't add any header row in the CSV file. The columns on each line need to be separated by a comma. Each cell can be wrapped in quotes which is helpful if a cell contains a comma.
A valid file could look like this:`,
		ImportHelp2: `hola,hello
"gracias","thanks, thank you"`,
		ImportPrompt:       "%d phrases have been detected. Would you like to import them into Slangbrain?",
		ImportPromptIgnore: "%d new phrases have been detected and %d have been becuase you already have them in your Slangbrain. Would you like to import the new phrases?",
		ImportNone:         "%d phrases have been detected, but you already have all of them in your Slangbrain.",
		ImportConfirm:      "%d phrases have been imported.",
		ImportCancel:       "Ok, no phrases have been imported.",
		ImportEmpty:        "The CSV file is empty. Nothing has been imported.",
		ImportErrParse: `The file '%s' is not formatted correctly. Please check the file and try it again. Parsing the file failed with the error:
'%v'`,
		ImportErrCols:      "Expecting CSV files to have 2 columns, but file '%s' has %d. The first one should contain the phrase, the second an explanation.",
		ImportErrDuplicate: "There are multiple phrases with the explantion '%s'. Please solve the conflict and try again.",
		AnHour:             "an hour",
		Hours:              "hours",
		AMinute:            "a minute",
		Minutes:            "minutes",
	}

	l := labels{
		StudyDone:            "done studying",
		Study:                "study",
		Add:                  "phrases",
		Done:                 "done",
		Help:                 "help",
		ImportHelp:           "import from CSV",
		Export:               "export as CSV",
		CloseImportHelp:      "ok",
		SubscribeConfirm:     "sounds good",
		SubscribeDeny:        "no thanks",
		DisableNotifications: "stop notifications",
		EnableNotifications:  "notify me",
		SendFeedback:         "send feedback",
		QuitHelp:             "all good",
		CancelFeedback:       "cancel",
		StopAdding:           "stop adding",
		ShowPhrase:           "show phrase",
		ScoreBad:             "didn't know",
		ScoreGood:            "got it",
		StudyNotNow:          "not now",
		ConfirmDelete:        "delete phrase",
		CancelDelete:         "cancel",
		Manage:               "manage phrases",
		ConfirmImport:        "yes",
		CancelImport:         "cancel",
	}

	w := Web{
		Title:         "Manage phrases",
		Search:        "Search",
		Empty:         "No phrases found.",
		Phrases:       "phrases",
		Phrase:        "Phrase",
		Explanation:   "Explanation",
		Delete:        "delete",
		Cancel:        "cancel",
		DeleteConfirm: "confirm delete",
		Save:          "save",
		Error:         "Something went wrong. Please try again.",
		Updated:       "updated phrase",
		Deleted:       "deleted phrase",
	}

	return m, l, w
}
