package messenger

// Defines the different sizes available when setting up a CallToActionsItem
// of type "web_url". These values can be used in the "WebviewHeightRatio"
// field.
const (
	// WebviewCompact opens the page in a web view that takes half the screen
	// and covers only part of the conversation.
	WebviewCompact = "compact"

	// WebviewTall opens the page in a web view that covers about 75% of the
	// conversation.
	WebviewTall = "tall"

	// WebviewFull opens the page in a web view that completely covers the
	// conversation, and has a "back" button instead of a "close" one.
	WebviewFull = "full"
)

// Greeting is a localized greeting message
type Greeting struct {
	Locale string `json:"locale"`
	Text   string `json:"text"`
}

// LocalizedMenu represents the menu for one language
type LocalizedMenu struct {
	Locale                string     `json:"locale,omitempty"`
	ComposerInputDisabled bool       `json:"composer_input_disabled,omitempty"`
	Items                 []MenuItem `json:"call_to_actions,omitempty"`
}

// MenuItem describes a single menu item
type MenuItem struct {
	Type    string     `json:"type"`
	Title   string     `json:"title"`
	Items   []MenuItem `json:"call_to_actions,omitempty"`
	Payload string     `json:"payload,omitempty"`
	URL     string     `json:"url,omitempty"`
	// One of WebviewCompact, WebviewTall, WebviewFull
	WebviewHeightRatio string `json:"webview_height_ratio,omitempty"`
	MessengerExtension bool   `json:"messenger_extensions,omitempty"`
	FallbackURL        string `json:"fallback_url,omitempty"`
	// Set to "hide" to disable sharing in the webview (for sensitive info).
	WebviewShareButton string `json:"webview_share_button,omitempty"`
}

type greetingSettings struct {
	Greeting []Greeting `json:"greeting,omitempty"`
}

type menuSettings struct {
	PersistentMenu []LocalizedMenu `json:"persistent_menu,omitempty"`
}

type getStartedSettings struct {
	GetStarted getStartedPayload `json:"get_started,omitempty"`
}

type getStartedPayload struct {
	Payload string `json:"payload,omitempty"`
}
