package payment

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/alissoncorsair/appsolidario-backend/payment"
	"github.com/alissoncorsair/appsolidario-backend/service/mailer"
	"github.com/alissoncorsair/appsolidario-backend/service/notification"
	"github.com/alissoncorsair/appsolidario-backend/service/transactions"
	"github.com/alissoncorsair/appsolidario-backend/service/user"
	"github.com/alissoncorsair/appsolidario-backend/types"
)

type Store struct {
	db                *sql.DB
	transactionsStore *transactions.Store
	userStore         *user.Store
	notificationStore *notification.Store
	gateway           payment.MercadoPago
	mailer            mailer.Mailer
}

func NewStore(db *sql.DB, gateway payment.MercadoPago, transactionsStore *transactions.Store, userStore *user.Store, notificationsStore *notification.Store, mailer mailer.Mailer) *Store {
	return &Store{
		db:                db,
		transactionsStore: transactionsStore,
		notificationStore: notificationsStore,
		userStore:         userStore,
		gateway:           gateway,
		mailer:            mailer,
	}
}

type CreatePaymentResponse struct {
	ExternalID    string  `json:"external_id"`
	QRCodeBase64  string  `json:"qr_code"`
	CopyPasteCode string  `json:"copy_paste_code"`
	Amount        float64 `json:"amount"`
}

func (s *Store) CreatePayment(paymentInfo payment.PaymentInfo, user types.User) (*CreatePaymentResponse, error) {
	info, err := s.gateway.GeneratePixPayment(paymentInfo, user)

	if err != nil {
		return nil, err
	}

	stringId := strconv.Itoa(info.ID)

	transaction, err := s.transactionsStore.GetTransactionByExternalID(stringId)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if transaction != nil {
		return &CreatePaymentResponse{
			ExternalID:    stringId,
			QRCodeBase64:  info.PointOfInteraction.TransactionData.QRCodeBase64,
			Amount:        info.TransactionAmount,
			CopyPasteCode: info.PointOfInteraction.TransactionData.QRCode,
		}, nil
	}

	_, err = s.transactionsStore.CreateTransaction(stringId, user.ID, paymentInfo.ReceiverID, paymentInfo.Amount, "Payment")

	if err != nil {
		return nil, err
	}

	response := &CreatePaymentResponse{
		ExternalID:    stringId,
		QRCodeBase64:  info.PointOfInteraction.TransactionData.QRCodeBase64,
		Amount:        info.TransactionAmount,
		CopyPasteCode: info.PointOfInteraction.TransactionData.QRCode,
	}

	return response, nil
}

type PaymentStatusResponse struct {
	Status        types.TransactionStatus `json:"status"`
	Amount        float64                 `json:"amount"`
	QRCode        string                  `json:"qr_code"`
	CopyPasteCode string                  `json:"copy_paste_code"`
}

func (s *Store) GetPaymentStatus(paymentID string) (*PaymentStatusResponse, error) {
	paymentInfo, err := s.gateway.GetPaymentStatus(paymentID)

	var status types.TransactionStatus = types.StatusPending
	if err != nil {
		return nil, err
	}

	if paymentInfo.Status == payment.MercadoPagoStatusApproved {
		_, err := s.transactionsStore.GetTransactionByExternalID(strconv.Itoa(paymentInfo.ID))

		if err != nil {
			return nil, err
		}

		status = types.StatusDone
		_, err = s.transactionsStore.UpdateTransactionStatusAndAmount(strconv.Itoa(paymentInfo.ID), types.StatusDone, paymentInfo.TransactionAmount)
		if err != nil {
			return nil, err
		}
	}

	qrCode := ""
	if status != types.StatusDone {
		qrCode = paymentInfo.PointOfInteraction.TransactionData.QRCodeBase64
	}

	response := &PaymentStatusResponse{
		Status:        status,
		Amount:        paymentInfo.TransactionAmount,
		QRCode:        qrCode,
		CopyPasteCode: paymentInfo.PointOfInteraction.TransactionData.QRCode,
	}

	return response, nil
}

func (s *Store) ProcessWebhookEvent(event payment.MercadoPagoWebhookEvent) error {
	switch event.Type {
	case "payment":
		paymentID := event.Data.ID
		if event.Action == payment.MercadoPagoWebhookActionPaymentCreated {
			return nil
		}

		paymentInfo, err := s.gateway.GetPaymentStatus(paymentID)

		if err != nil {
			return fmt.Errorf("failed to get payment status: %w", err)
		}

		if paymentInfo.Status == payment.MercadoPagoStatusApproved {
			transaction, err := s.transactionsStore.GetTransactionByExternalID(paymentID)

			if err != nil {
				return err
			}

			if transaction == nil {
				return fmt.Errorf("transaction not found")
			}

			_, err = s.transactionsStore.UpdateTransactionStatusAndAmount(paymentID, types.StatusDone, paymentInfo.TransactionAmount)

			if err != nil {
				return err
			}

			notification := &types.Notification{
				UserID:     transaction.PayeeID,
				FromUserID: transaction.PayerID,
				Type:       types.TypePayment,
				ResourceID: transaction.ID,
				IsRead:     false,
			}

			_, _ = s.notificationStore.CreateNotification(notification)

			payer, err := s.userStore.GetUserByID(transaction.PayerID)
			if err == nil && payer != nil {
				err = s.mailer.SendPaymentThanksEmail(payer, paymentInfo.TransactionAmount)
				if err != nil {
					fmt.Printf("failed to send payment thanks email: %v", err)
				}
			}
		}
	default:
		return fmt.Errorf("unhandled event type: %s", event.Type)
	}

	return nil
}
