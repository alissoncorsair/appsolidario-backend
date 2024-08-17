package mailer

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/alissoncorsair/appsolidario-backend/types"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var FromName = "Solidariza"

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

func (m *SendGridMailer) SendConfirmationEmail(user *types.User, token string) error {
	from := mail.NewEmail(FromName, m.From)
	subject := "Confirmação de e-mail"
	userName := fmt.Sprintf("%s %s", user.Name, user.Surname)
	to := mail.NewEmail(userName, user.Email)

	body, err := BuildConfirmationEmail(user, token)

	if err != nil {
		return err
	}

	message := mail.NewSingleEmail(from, subject, to, "", body)

	_, err = m.Client.Send(message)

	return err
}

func BuildConfirmationEmail(user *types.User, token string) (string, error) {
	templ, err := template.ParseFiles("service/mailer/mail-confirmation.templ")
	if err != nil {
		return "", err
	}

	URL := fmt.Sprintf("http://localhost:8080/confirm-email?token=%s", token)

	payload := struct {
		User *types.User
		URL  string
	}{
		User: user,
		URL:  URL,
	}

	var body bytes.Buffer
	err = templ.Execute(&body, payload)

	if err != nil {
		return "", err
	}

	return body.String(), nil
}
