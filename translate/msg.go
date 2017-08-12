package translate

// Msg contains all messages ever send to a user.
type Msg struct {
	// Greeting is currently limited to 160 chars
	Greeting,
	Menu,
	Help,
	Idle,
	Add,
	Welcome1,
	Welcome2,
	Error,
	ExplanationMissing,
	PhraseMissing,
	StudyDone,
	StudyCorrect,
	StudyWrong,
	StudyEmpty,
	StudyQuestion,
	ExplanationExists,
	AddDone,
	AddNext,
	StudyNotification,
	ConfirmDelete,
	Deleted,
	CancelDelete,
	AskToSubscribe,
	Subscribed,
	ConfirmUnsubscribe,
	Unsubscribed,
	Fedback,
	FeedbackDone string
}
