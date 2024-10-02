package payment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/alissoncorsair/appsolidario-backend/types"
	"github.com/alissoncorsair/appsolidario-backend/utils"
)

var url = "https://api.mercadopago.com/v1/payments"

type MercadoPago struct {
	AccessToken string
}

type PaymentInfo struct {
	Amount         float64 `json:"amount"`
	Description    string  `json:"description"`
	ReceiverID     int     `json:"receiver_id"`
	IdempotencyKey string  `json:"idempotency_key"`
}

type Identification struct {
	Type   string `json:"type"`
	Number string `json:"number"`
}

type Address struct {
	ZIPCode     string `json:"zip_code"`
	City        string `json:"city"`
	FederalUnit string `json:"federal_unit"`
}

type Payer struct {
	Email          string         `json:"email"`
	FirstName      string         `json:"first_name"`
	LastName       string         `json:"last_name"`
	Identification Identification `json:"identification"`
	Address        Address        `json:"address"`
}

type GeneratePixPaymentRequest struct {
	TransactionAmount float64 `json:"transaction_amount"`
	Description       string  `json:"description"`
	PaymentMethodId   string  `json:"payment_method_id"`
	Payer             Payer   `json:"payer"`
}

type TransactionDetails struct {
	NetReceivedAmount float64 `json:"net_received_amount"`
	TotalPaidAmount   float64 `json:"total_paid_amount"`
	InstallmentAmount float64 `json:"installment_amount"`
}

type FeeDetail struct {
	Type     string  `json:"type"`
	Amount   float64 `json:"amount"`
	FeePayer string  `json:"fee_payer"`
}

type PayerResponse struct {
	ID             string         `json:"id"`
	Email          string         `json:"email"`
	Identification Identification `json:"identification"`
	Type           string         `json:"type"`
}

type PointOfInteraction struct {
	Type            string          `json:"type"`
	TransactionData TransactionData `json:"transaction_data"`
}

type TransactionData struct {
	QRCode       string `json:"qr_code"`
	QRCodeBase64 string `json:"qr_code_base64"`
	TicketURL    string `json:"ticket_url"`
}

type MercadoPagoPixResponse struct {
	ID                 int                `json:"id"`
	Status             string             `json:"status"`
	PaymentTypeID      string             `json:"payment_type_id"`
	TransactionAmount  float64            `json:"transaction_amount"`
	CurrencyID         string             `json:"currency_id"`
	DateApproved       string             `json:"date_approved"`
	DateCreated        string             `json:"date_created"`
	DateLastUpdated    string             `json:"date_last_updated"`
	MoneyReleaseDate   string             `json:"money_release_date"`
	Description        string             `json:"description"`
	Payer              PayerResponse      `json:"payer"`
	ExternalReference  string             `json:"external_reference"`
	TransactionDetails TransactionDetails `json:"transaction_details"`
	FeeDetails         []FeeDetail        `json:"fee_details"`
	PointOfInteraction PointOfInteraction `json:"point_of_interaction"`
}

type MercadoPagoStatusResponse string

const (
	MercadoPagoStatusPending  MercadoPagoStatusResponse = "pending"
	MercadoPagoStatusApproved MercadoPagoStatusResponse = "approved"
)

type MercadoPagoPaymentStatusResponse struct {
	ID                 int                       `json:"id"`
	Status             MercadoPagoStatusResponse `json:"status"`
	PointOfInteraction PointOfInteraction        `json:"point_of_interaction"`
	TransactionAmount  float64                   `json:"transaction_amount"`
}

func (mp *MercadoPago) GeneratePixPayment(paymentInfo PaymentInfo, user types.User) (*MercadoPagoPixResponse, error) {
	jsonStr := GeneratePixPaymentRequest{
		TransactionAmount: paymentInfo.Amount,
		Description:       paymentInfo.Description,
		PaymentMethodId:   "pix",
		Payer: Payer{
			Email:     user.Email,
			FirstName: user.Name,
			LastName:  user.Surname,
			Identification: Identification{
				Type:   "CPF",
				Number: user.CPF,
			},
			Address: Address{
				ZIPCode:     user.PostalCode,
				City:        user.City,
				FederalUnit: user.State,
			},
		},
	}
	client := utils.GetHttpClient()
	marshalled, err := json.Marshal(jsonStr)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(marshalled))

	if err != nil {
		return nil, err
	}

	req.Header = http.Header{
		"Authorization":     []string{"Bearer " + mp.AccessToken},
		"Content-Type":      []string{"application/json"},
		"X-Idempotency-Key": []string{paymentInfo.IdempotencyKey},
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var mpResp MercadoPagoPixResponse

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	err = json.NewDecoder(bytes.NewReader(body)).Decode(&mpResp)

	if err != nil {
		return nil, fmt.Errorf("error decoding JSON response: %w, body: %s", err, string(body))
	}

	return &mpResp, nil
}

func (mp *MercadoPago) GetPaymentStatus(paymentID string) (*MercadoPagoPaymentStatusResponse, error) {
	client := utils.GetHttpClient()
	req, err := http.NewRequest("GET", url+"/"+paymentID, nil)

	if err != nil {
		return nil, err
	}

	req.Header = http.Header{
		"Authorization": []string{"Bearer " + mp.AccessToken},
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var mpResp MercadoPagoPaymentStatusResponse

	err = json.NewDecoder(resp.Body).Decode(&mpResp)

	if err != nil {
		return nil, err
	}

	return &mpResp, nil
}
