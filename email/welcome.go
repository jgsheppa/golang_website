package email

import (
	"fmt"
	"os"

	"github.com/mailgun/mailgun-go"
)

func SendWelcomeEmail() error {
	domain := os.Getenv("MAILGUN_DOMAIN")
	apiKey := os.Getenv("MAILGUN_PRIVATE_KEY")
	
	mg := mailgun.NewMailgun(domain, apiKey)
	m := mg.NewMessage(
		"Excited User <jgsheppa@protonmail.com>",
		"Hello",
		"Testing some Mailgun awesomeness!",
		"jgsheppard92@gmail.com",
	)
	_, id, err := mg.Send(m)
	fmt.Println("ID", id)
	return err
}