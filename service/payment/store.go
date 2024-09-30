package payment

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/alissoncorsair/appsolidario-backend/payment"
	"github.com/alissoncorsair/appsolidario-backend/service/transactions"
	"github.com/alissoncorsair/appsolidario-backend/types"
)

type Store struct {
	db                *sql.DB
	transactionsStore *transactions.Store
	gateway           payment.MercadoPago
}

func NewStore(db *sql.DB, gateway payment.MercadoPago, transactionsStore *transactions.Store) *Store {
	return &Store{
		db:                db,
		transactionsStore: transactionsStore,
		gateway:           gateway,
	}
}

type CreatePaymentResponse struct {
	ExternalID   string `json:"external_id"`
	QRCodeBase64 string `json:"qr_code"`
	Amount       int    `json:"amount"`
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
			ExternalID:   stringId,
			QRCodeBase64: info.PointOfInteraction.TransactionData.QRCodeBase64,
			Amount:       int(info.TransactionAmount),
		}, nil
	}

	_, err = s.transactionsStore.CreateTransaction(stringId, user.ID, paymentInfo.ReceiverID, paymentInfo.Amount, "Payment")

	if err != nil {
		return nil, err
	}

	response := &CreatePaymentResponse{
		ExternalID:   stringId,
		QRCodeBase64: info.PointOfInteraction.TransactionData.QRCodeBase64,
		Amount:       int(info.TransactionAmount),
	}

	return response, nil
}

type PaymentStatusResponse struct {
	Status types.TransactionStatus `json:"status"`
	QRCode string                  `json:"qr_code"`
}

func (s *Store) GetPaymentStatus(paymentID string) (*PaymentStatusResponse, error) {
	paymentInfo, err := s.gateway.GetPaymentStatus(paymentID)

	var status types.TransactionStatus = types.StatusPending
	if err != nil {
		return nil, err
	}

	if paymentInfo.Status == payment.MercadoPagoStatusApproved {
		status = types.StatusDone
		paymentID := strconv.Itoa(paymentInfo.ID)
		_, err = s.transactionsStore.UpdateTransactionStatus(paymentID, types.StatusDone)
		fmt.Println(err)
		if err != nil {
			return nil, err
		}
	}

	qrCode := ""
	if status != types.StatusDone {
		qrCode = paymentInfo.PointOfInteraction.TransactionData.QRCodeBase64
	}

	response := &PaymentStatusResponse{
		Status: status,
		QRCode: qrCode,
	}

	return response, nil
}
