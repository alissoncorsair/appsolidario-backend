package mailer

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/alissoncorsair/appsolidario-backend/config"
	"github.com/alissoncorsair/appsolidario-backend/types"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var FromName = "Solidariza"

type SendGridMailer struct {
	From   string
	Client *sendgrid.Client
}

func NewSendGridMailer(apiKey, fromEmail string) *SendGridMailer {
	return &SendGridMailer{
		From:   fromEmail,
		Client: sendgrid.NewSendClient(apiKey),
	}
}

func (m *SendGridMailer) SendConfirmationEmail(user *types.User, token string) error {
	from := mail.NewEmail(FromName, m.From)
	subject := "Confirmação de e-mail"
	userName := fmt.Sprintf("%s %s", user.Name, user.Surname)
	to := mail.NewEmail(userName, user.Email)

	plainTextContent, err := BuildConfirmationEmail(user, token)
	if err != nil {
		return fmt.Errorf("failed to build confirmation email: %w", err)
	}

	htmlContent, err := BuildConfirmationEmail(user, token)
	if err != nil {
		return fmt.Errorf("failed to build confirmation email: %w", err)
	}

	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)

	response, err := m.Client.Send(message)

	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("failed to send email: %s", response.Body)
	}

	fmt.Printf("Email sent. Status Code: %d, Body: %s, Headers: %v\n", response.StatusCode, response.Body, response.Headers)

	return nil
}

func BuildConfirmationEmail(user *types.User, token string) (string, error) {
	templ, err := template.ParseFiles("service/mailer/mail-confirmation.templ")
	if err != nil {
		return "", err
	}

	URL := fmt.Sprintf("%s?token=%s", config.Envs.EmailVerifyUrl, token)

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
