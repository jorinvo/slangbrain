package messenger

const (
	messageStartMenu = `What would you like to do next?
Please use the buttons below.`
	messageHelp = `this is help
`
	messageIdle     = "Good, just send me a \U0001F44D to continue with your studies."
	messageStartAdd = `Please send me a phrase and its explanation.
Separate them with a linebreak.`
	messageWelcome = `Hello %s!
Slangebrain is here to help you with your language studies.
Whenever you pick up a new phrase, just add it to your Slangebrain and remember it forever.
Master the language you encounter in your every day life instead of being limited to a textbook.`
	messageWelcome2 = `You begin by adding phrases and after, Slangbrain will test your memories in a natural schedule.

` + messageStartAdd
	messageErr              = "Sorry, something went wrong."
	messageExplanationEmpty = "The phrase is missing an explanation. Please send it again with explanation."
	messagePhraseEmpty      = "Please send a phrase."
	messageStudyDone        = `Congrats, you finished all your studies for now!
Come back in %s.`
	messageStudyCorrect = "Nice, your answer was correct!"
	messageStudyWrong   = `Sorry, the correct version is:

%s

`
	messageStudyEmpty = `You have added no phrases yet.
Click the button below and get started.`
	messageStudyQuestion = `%d. Do you remember how to say this?

%s

Use the buttons or type the phrase.`
	messageExplanationExists = `You already saved a phrase with the same explanation:
%s
%s

Please send it again with an explanation you can distinguish from the existing one.`
	messageAddDone = `Saved phrase:
%s

With explanation:
%s`
	messageAddNext       = "Add next phrase."
	messageStudiesDue    = `Hey %s, there are %d phrases ready for review!`
	messageConfirmDelete = "Are you sure, you want to delete this phrase?"
	messageDeleted       = "The phrase has been deleted. Let's continue studying other phrases."
	messageCancelDelete  = "Good, let's keep that phrase and continue studying."
	greeting             = `Slangbrain helps you with our language Studies.
Master the language you encounter in your every day life instead of being limited to a textbook.`
)
