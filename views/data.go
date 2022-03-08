package views

import (
	"log"
	"net/http"
	"time"

	"github.com/jgsheppa/golang_website/models"
)

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
		log.Println(err)
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

func persistAlert(w http.ResponseWriter, alert Alert) {
	expiresAt := time.Now().Add(20 * time.Second)
	level := http.Cookie{
		Name: "alert_level",
		Value: alert.Level,
		Expires: expiresAt,
		HttpOnly: true,
	}
	message := http.Cookie{
		Name: "alert_message",
		Value: alert.Message,
		Expires: expiresAt,
		HttpOnly: true,
	}

	http.SetCookie(w, &level)
	http.SetCookie(w, &message)
}

func clearAlert(w http.ResponseWriter) {
	level := http.Cookie{
		Name: "alert_level",
		Value: "",
		Expires: time.Now(),
		HttpOnly: true,
	}
	message := http.Cookie{
		Name: "alert_message",
		Value: "",
		Expires: time.Now(),
		HttpOnly: true,
	}

	http.SetCookie(w, &level)
	http.SetCookie(w, &message)
}

func getAlert(r *http.Request) *Alert {
	level, err := r.Cookie("alert_level")
	if err != nil {
		return nil
	}
	message, err := r.Cookie("alert_message")
	if err != nil {
		return nil
	}
	alert := Alert{
		Level: level.Value,
		Message: message.Value,
	}
	return &alert
}

func RedirectAlert(w http.ResponseWriter, r *http.Request, urlStr string, code int, alert Alert) {
	persistAlert(w, alert)
	http.Redirect(w, r, urlStr, code)
}