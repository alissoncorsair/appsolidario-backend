package mailer

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/alissoncorsair/appsolidario-backend/config"
	"github.com/alissoncorsair/appsolidario-backend/types"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var FromName = "Solidariza"

type SendGridMailer struct {
	Client   *sendgrid.Client
	From     string
	FromName string
	DevMode  bool
}

func NewSendGridMailer(apiKey, fromEmail string, devMode bool) *SendGridMailer {
	return &SendGridMailer{
		From:    fromEmail,
		Client:  sendgrid.NewSendClient(apiKey),
		DevMode: devMode,
	}
}

func (m *SendGridMailer) SendConfirmationEmail(user *types.User, token string) error {
	if m.DevMode {
		fmt.Printf("Development mode: Email not sent. User: %s, Token: %s\n", user.Email, token)
		return nil
	}

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

func (m *SendGridMailer) SendPaymentThanksEmail(user *types.User, amount float64) error {
	/*     if m.DevMode {
	       fmt.Printf("Development mode: Thank you email would be sent to %s\n", user.Email)
	       return nil
	   } */

	from := mail.NewEmail(m.FromName, m.From)
	subject := "Obrigado pela sua doação!"
	to := mail.NewEmail(user.Name, user.Email)

	templateData := struct {
		User        *types.User
		Amount      float64
		CurrentYear int
	}{
		User:        user,
		Amount:      amount,
		CurrentYear: time.Now().Year(),
	}

	var body bytes.Buffer
	tmpl, err := template.ParseFiles("service/mailer/templates/payment-thanks.templ")
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	if err := tmpl.Execute(&body, templateData); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	message := mail.NewSingleEmail(from, subject, to, "", body.String())
	response, err := m.Client.Send(message)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("email API error: %d - %s", response.StatusCode, response.Body)
	}

	return nil
}

func BuildConfirmationEmail(user *types.User, token string) (string, error) {
	templ, err := template.ParseFiles("service/mailer/templates/mail-confirmation.templ")
	if err != nil {
		return "", err
	}

	URL := fmt.Sprintf("%s?token=%s", config.Envs.EmailVerifyUrl, token)

	payload := struct {
		User        *types.User
		URL         string
		CurrentYear int
	}{
		User:        user,
		URL:         URL,
		CurrentYear: time.Now().Year(),
	}

	var body bytes.Buffer
	err = templ.Execute(&body, payload)

	if err != nil {
		return "", err
	}

	return body.String(), nil
}
