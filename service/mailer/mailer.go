package mailer

import "github.com/alissoncorsair/appsolidario-backend/types"

type Mailer interface {
	SendConfirmationEmail(user *types.User, token string) error
}
