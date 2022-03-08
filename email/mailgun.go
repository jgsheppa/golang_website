package email

import (
	"fmt"

	"github.com/mailgun/mailgun-go"
)

const welcomeSubject = "Schnup Support"

const welcomeHTML = `
<html>
<body>
	<h1>Sending HTML emails with Mailgun</h1>
	<p>Dear User</p>
	<p style="font-size:30px;">More examples can be found <a href="https://documentation.mailgun.com/en/latest/api-sending.html#examples">here</a></p>
</body>
</html>
`

const welcomeMessage = "Hello from Schnup!"

func WithMailgun(domain, apiKey string) ClientConfig {
	return func(c *Client) {
		mg := mailgun.NewMailgun(domain, apiKey)
		c.mg = mg
	}
}

func WithSender(name, email string) ClientConfig {
	return func(c *Client){
		c.from = buildEmail(name, email)
	}
}

type ClientConfig func(*Client)

func NewClient(opts ...ClientConfig) *Client {
	client := Client{
		from: "jgsheppard92@gmail.com",
	}

	for _, opt := range opts {
		opt(&client)
	}
	return &client
}

type Client struct {
	from string
	mg mailgun.Mailgun
}

func (c *Client) Welcome(toName, toEmail string) error {
	recipient := buildEmail(toName, toEmail)
	message := mailgun.NewMessage(c.from, welcomeSubject, welcomeMessage, recipient)
	message.SetHtml(welcomeHTML)

	_, _, err := c.mg.Send(message)
	return err
}

func buildEmail(name, email string) string {
	if name == "" {
		return email
	}
	return fmt.Sprintf("%s <%s>", name, email)
}

