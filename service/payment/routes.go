package payment

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alissoncorsair/appsolidario-backend/payment"
	"github.com/alissoncorsair/appsolidario-backend/service/auth"
	"github.com/alissoncorsair/appsolidario-backend/service/user"
	"github.com/alissoncorsair/appsolidario-backend/utils"
)

type Handler struct {
	paymentStore *Store
	userStore    *user.Store
}

func NewHandler(paymentStore *Store, userStore *user.Store) *Handler {
	return &Handler{
		paymentStore: paymentStore,
		userStore:    userStore,
	}
}

func (h *Handler) HandleGeneratePix(w http.ResponseWriter, r *http.Request) {
	var payload payment.PaymentInfo
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("failed to decode request: %w", err))
		return
	}

	if err := utils.Validate.Struct(payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("user not authenticated"))
		return
	}

	user, err := h.userStore.GetUserByID(userID)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to get user: %w", err))
		return
	}

	payee, err := h.userStore.GetUserByID(payload.ReceiverID)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to get payee: %w", err))
		return
	}

	if payee.ID == user.ID {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("cannot pay yourself"))
		return
	}

	info, err := h.paymentStore.CreatePayment(payload, *user)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to create payment: %w", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, info)
}

func (h *Handler) HandleGetPaymentStatus(w http.ResponseWriter, r *http.Request) {
	paymentID := r.PathValue("payment_id")

	info, err := h.paymentStore.GetPaymentStatus(paymentID)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to get payment status: %w", err))
		return
	}

	response := map[string]interface{}{
		"data": info,
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) HandleMercadoPagoWebhook(w http.ResponseWriter, r *http.Request) {
	var webhookEvent payment.MercadoPagoWebhookEvent
	err := json.NewDecoder(r.Body).Decode(&webhookEvent)

	fmt.Println("webhookEvent", webhookEvent)

	if err != nil {
		fmt.Println("err decoding", err)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("failed to decode request: %w", err))
		return
	}

	err = h.paymentStore.ProcessWebhookEvent(webhookEvent)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to handle webhook: %w", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}

func (h *Handler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("POST /pay", auth.WithJWTAuth(h.HandleGeneratePix, h.userStore))
	router.HandleFunc("GET /pay/status/{payment_id}", auth.WithJWTAuth(h.HandleGetPaymentStatus, h.userStore))
	router.HandleFunc("POST /webhook/mpago", h.HandleMercadoPagoWebhook)
}
