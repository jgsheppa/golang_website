package views

const (
	AlertLevelDanger = "danger"
	AlertLevelWarning = "warning"
	AlertLevelInfo = "info"
	AlertLevelSuccess = "success"

	AlertMsgGeneric = "Something went wrong. Please try again. If the problem persists, please contact us."
)

// User to render bootstrap alert messages
// in the user interface
type Alert struct {
	Level string
	Message string
}

// This is used to pass dynamic data to HTML templates
type Data struct {
	Alert *Alert
	Yield interface{}
}