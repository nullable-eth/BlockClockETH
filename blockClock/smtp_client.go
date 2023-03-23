package blockClock

import (
	"fmt"
	"net/smtp"
)

// EmailClient - sends emails using generic smtp server.
type EmailClient struct {
	auth          smtp.Auth
	server        string
	port          string
	from          string
	NotifyAddress string
}

// NewEmailClient creates and returns a smtp client to use for email sending.
func NewEmailClient(server string, port string, user string, password string) (*EmailClient, error) {
	c := EmailClient{
		auth:   smtp.PlainAuth("", user, password, server),
		server: server,
		port:   port,
		from:   user,
	}

	return &c, nil
}

// SendEmail sends a templates email.
func (c *EmailClient) SendEmail(body string, subject string, to string) error {
	var err error

	msg := c.buildBody(c.from, to, subject, body)

	// must be over an SSL or TLS connection or this will fail because SendMail refuses to send credentials over
	// un-encrypted connections.
	err = smtp.SendMail(c.server+":"+c.port, c.auth, c.from, []string{to}, []byte(msg))

	return err
}

func (c *EmailClient) buildBody(from string, to string, subject string, body string) string {
	msg := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\r\n"
	msg += fmt.Sprintf("From: %s\r\n", from)
	msg += fmt.Sprintf("To: %s\r\n", to)
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += fmt.Sprintf("\r\n%s\r\n", body)
	return msg
}
