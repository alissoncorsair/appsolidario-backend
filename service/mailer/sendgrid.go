package mailer

import "github.com/sendgrid/sendgrid-go"

type SendGridMailer struct {
	From   string
	Client *sendgrid.Client
}

func NewSendGridMailer(from, apiKey string) *SendGridMailer {
	return &SendGridMailer{
		From:   from,
		Client: sendgrid.NewSendClient(apiKey),
	}
}
