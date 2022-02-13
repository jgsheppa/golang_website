package views

import "github.com/jgsheppa/golang_website/models"

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
	User *models.User
	Yield interface{}
}

func (d *Data) SetAlert(err error) {
	if pErr, ok := err.(PublicError); ok {
		d.Alert = &Alert{
			Level: AlertLevelDanger,
			Message: pErr.Public(),
		}
	} else {
		d.Alert = &Alert{
			Level: AlertLevelDanger,
			Message: AlertMsgGeneric,
		}
	}
}

func (d *Data) AlertError(msg string) {
	d.Alert = &Alert{
		Level: AlertLevelDanger,
		Message: msg,
	}
}

type PublicError interface {
	error 
	Public() string
}