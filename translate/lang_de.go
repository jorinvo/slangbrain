package translate

// de provides content in German.
func de() (Msg, buttonLabels) {
	m := Msg{
		// Greeting is currently limited to 160 chars.
		// {{user_first_name}} is replaced by Facebook.
		Greeting: `Hi {{user_first_name}}! Mit Slangbrain kannst du Sprache lernen wie sie dir im Alltag begegnet statt ein Schulbuch auswendig zu lernen.`,
		Menu:     `Was willst du als nächstes machen?`,
		Help:     "Wie kann ich dir weiterhelfen?",
		Idle:     "Alles klar. Schicke mir einfach ein " + iconThumbsup + " um weiterzumachen.",
		Add:      `Schicke ein Wort oder einen Satz in der Sprache, die du lernst und nach einer leeren Zeile kannst du eine Erklärung in Deutsch hinzufügen.`,
		Welcome1: `Hallo %s!

Jedes Mal wenn du ein neues Wort im Alltag lernst, füge es einfach zu Slangbrain hinzu und vergesse es nie wieder.

Nachdem du Vokabeln gespeichert hast, wird Slangbrain dich automatisch in sinnvollen Abständen abfragen und du wirst dich immer an die Wörter erinnern.`,
		Welcome2: `Bitte schicke jetzt einen Satz in der Sprache die du lernst und nach einer leeren Zeile kannst du eine Erklärung in Deutsch hinzufügen.

Ein Beispiel wäre, wenn du Französisch lernst, dann könntest du folgende Nachricht schicken:`,
		Welcome3: `Bonjour !
Guten Tag!`,
		Welcome4:           "Jetzt bist du dran:",
		Error:              "Entschuldigung, etwas ist schief gelaufen.",
		ExplanationMissing: "Jede Vokabel braucht auch eine Erklärung. Schicke nochmal eine Nachricht.",
		PhraseMissing:      "Schicke eine Vokabel.",
		StudyDone: `Glückwunsch, du hast alle Vokabeln fürs Erste geschafft!
In %s gibt es wieder etwas zu wiederholen.`,
		StudyCorrect: "Richtig!",
		StudyWrong: `Richtig heißt es:

%s`,
		StudyEmpty: `Du hast noch keine Vokabeln hinzugefügt.
Klicke den Button um anzufangen:`,
		StudyQuestion: `%d. Kannst du das übersetzen?

%s

Schicke mir die Antwort oder klicke den Button.`,
		ExplanationExists: `Du hast schon eine Vokabel mit der gleichen Erklärung:
%s
%s

Bitte sende nochmal eine Nachricht mit einer Erklärung die sich von der vorhandenen unterscheidet.`,
		AddDone: `Gespeichert:
%s

Mit Erklärung:
%s`,
		AddNext:            "Schicke die nächste Vokabel.",
		StudyNotification:  `Hey %s, %d Vokabeln warten auf dich!`,
		ConfirmDelete:      "Bist du dir sicher, dass du die Vokabel löschen möchtest?",
		Deleted:            "Ist gelöscht. Jetzt geht es weiter mit lernen.",
		CancelDelete:       "Alles klar, dann geht es normal weiter mit lernen.",
		AskToSubscribe:     `Möchtest du, dass ich dir eine Nachricht schicke sobald es Vokabeln zu wiederholen gibt?`,
		Subscribed:         `Ich schicke dir eine Nachricht sobald es etwas zu wiederholen gibt.`,
		ConfirmUnsubscribe: `Alles klar, du bekommst in Zukunft keine Benachrichtigungen mehr.`,
		DenySubscribe:      `Alles klar.`,
		Feedback:           "Ein Problem ist aufgetreten, du hast einen Verbesserungsvorschlag für uns oder du willst einfach nur hallo sagen? Sende jetzt eine Nachricht und sie wird weitergeleitet an die Menschen die Slangbrain entschickeln.",
		FeedbackDone:       "Danke %s, wir melden uns bei dir sobald wie möglich.",
		AnHour:             "einer Stunde",
		Hours:              "Stunden",
		AMinute:            "einer Minute",
		Minutes:            "Minuten",
	}

	b := buttonLabels{
		StudyDone:         "genug gelernt",
		Study:             "lernen",
		Add:               "neu",
		Done:              "fertig",
		Help:              "Hilfe",
		SubscribeConfirm:  "gerne",
		SubscribeDeny:     "nein",
		StopNotifications: iconDelete + " Benachrichtigung",
		SendFeedback:      "Feedback geben",
		QuitHelp:          "zurück",
		CancelFeedback:    "abbrechen",
		StopAdding:        "stop",
		ShowPhrase:        "zeigen",
		ScoreBad:          "weiß nicht",
		ScoreGood:         "weiß ich",
		StudyNotNow:       "nicht jetzt",
		ConfirmDelete:     "jetzt löschen",
		CancelDelete:      "abbrechen",
	}

	return m, b
}
