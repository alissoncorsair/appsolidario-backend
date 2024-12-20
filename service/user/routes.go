package user

import (
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/alissoncorsair/appsolidario-backend/config"
	"github.com/alissoncorsair/appsolidario-backend/service/auth"
	"github.com/alissoncorsair/appsolidario-backend/service/mailer"
	"github.com/alissoncorsair/appsolidario-backend/service/notification"
	"github.com/alissoncorsair/appsolidario-backend/service/profile_picture"
	"github.com/alissoncorsair/appsolidario-backend/storage"
	"github.com/alissoncorsair/appsolidario-backend/types"
	"github.com/alissoncorsair/appsolidario-backend/utils"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/paemuri/brdoc"
)

type Handler struct {
	userStore         *Store
	pictureStore      *profile_picture.Store
	notificationStore *notification.Store
	storage           *storage.R2Storage
	mailer            mailer.Mailer
}

func NewHandler(userStore *Store, pictureStore *profile_picture.Store, notificationStore *notification.Store, storage *storage.R2Storage, mailer mailer.Mailer) *Handler {
	return &Handler{
		userStore:         userStore,
		pictureStore:      pictureStore,
		notificationStore: notificationStore,
		storage:           storage,
		mailer:            mailer,
	}
}

func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var payload types.RegisterUserRequest

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	existingUser, err := h.userStore.GetUserByEmail(payload.Email)

	if existingUser != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("o email %s já foi cadastrado", payload.Email))
		return
	}

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	var cpf string
	re := regexp.MustCompile("[^0-9]")
	cpf = re.ReplaceAllString(payload.CPF, "")

	valid := brdoc.IsCPF(cpf)

	if !valid {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("cpf inválido"))
		return
	}

	existingUser, err = h.userStore.GetUserByCPF(cpf)

	if existingUser != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("o cpf %s já foi cadastrado", payload.CPF))
		return
	}

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	hashedPass, err := auth.HashPassword(payload.Password)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	var birthDate time.Time

	birthDate, err = utils.ParseDate(payload.BirthDate)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("invalid birth date"))
		return
	}

	var roleID int
	roleID, err = utils.GetInt(payload.RoleID)

	if roleID != int(types.RolePayee) && roleID != int(types.RolePayer) {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid role id"))
		return
	}

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	postalCode := re.ReplaceAllString(payload.PostalCode, "")

	user, err := h.userStore.CreateUser(&types.User{
		UserWithoutPassword: types.UserWithoutPassword{
			Name:       payload.Name,
			Surname:    payload.Surname,
			Email:      payload.Email,
			PostalCode: postalCode,
			State:      payload.State,
			City:       payload.City,
			Status:     types.StatusInactive,
			RoleID:     types.UserRole(roleID),
			CPF:        cpf,
			BirthDate:  birthDate,
		},
		Password: hashedPass,
	})

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	activationToken, err := auth.CreateJWT([]byte(config.Envs.JWTSecret), user.ID, types.TokenTypeVerify, time.Hour*24)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.mailer.SendConfirmationEmail(user, activationToken)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to send confirmation email: %w", err))
		return
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]string{"message": "Usuário criado com sucesso, verifique seu email para ativar sua conta"})
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var payload types.LoginUserRequest

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	var cpf string
	re := regexp.MustCompile("[^0-9]")
	cpf = re.ReplaceAllString(payload.CPF, "")

	user, err := h.userStore.GetUserByCPF(cpf)

	if err != nil {
		//should not tell if the user exists or not
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("CPF/Senha inválidos"))
		return
	}

	if user == nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("CPF/Senha inválidos"))
		return
	}

	if !auth.ComparePassword(user.Password, []byte(payload.Password)) {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("CPF/Senha inválidos"))
		return
	}

	if user.Status == types.StatusInactive {
		utils.WriteError(w, http.StatusForbidden, fmt.Errorf("email not verified"))
		return
	}

	accessToken, err := auth.CreateJWT([]byte(config.Envs.JWTSecret), user.ID, types.TokenTypeAccess, time.Duration(config.Envs.JWTExpirationInSeconds)*time.Second)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	refreshToken, err := auth.CreateJWT([]byte(config.Envs.JWTSecret), user.ID, types.TokenTypeRefresh, time.Duration(config.Envs.JWTRefreshExpirationInSeconds)*time.Second)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	response := struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) HandleVerify(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	if token == "" {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("token is required"))
		return
	}

	validatedToken, err := auth.ValidateToken(token, types.TokenTypeVerify)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid token"))
		return
	}

	claims, ok := validatedToken.Claims.(jwt.MapClaims)
	if !ok {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("invalid token claims"))
		return
	}

	userIDStr, ok := claims["userID"].(string)
	if !ok {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("invalid user ID in token"))
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("invalid user ID format"))
		return
	}

	user, err := h.userStore.GetUserByID(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error getting user from id %d: %w", userID, err))
		return
	}

	if user.Status == types.StatusActive {
		utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Email already verified"})
		return
	}

	err = h.userStore.UpdateUserStatus(userID, types.StatusActive)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Email verified successfully"})
}

func (h *Handler) HandleAddProfilePicture(w http.ResponseWriter, r *http.Request) {
	userId, found := auth.GetUserIDFromContext(r.Context())

	if !found {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("user not found"))
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10 MB

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("failed to parse form: %w", err))
		return
	}

	file, fileHeader, err := r.FormFile("profile_picture")

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("failed to get file: %w", err))
		return
	}

	defer file.Close()

	_, filename, err := h.storage.UploadFile(r.Context(), file, fileHeader.Filename)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to upload file: %w", err))
		return
	}

	profilePicture, err := h.userStore.GetUserProfilePicture(userId)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to get profile picture: %w", err))
		return
	}

	if profilePicture != nil {
		err = h.storage.DeleteFile(r.Context(), profilePicture.Path)

		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to delete old profile picture: %w", err))
			return
		}
		newProfilePicture := &types.ProfilePicture{
			ID:     profilePicture.ID,
			UserID: userId,
			Path:   filename,
		}

		h.pictureStore.UpdateProfilePicture(newProfilePicture)
		utils.WriteJSON(w, http.StatusOK, newProfilePicture)
		return
	}

	profilePicture = &types.ProfilePicture{
		UserID: userId,
		Path:   filename,
	}

	profilePicture, err = h.pictureStore.CreateProfilePicture(profilePicture)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to create profile picture: %w", err))
		return
	}

	utils.WriteJSON(w, http.StatusCreated, profilePicture)
}

func (h *Handler) HandleGetProfile(id int, w http.ResponseWriter, r *http.Request) {
	user, err := h.userStore.GetUserByID(id)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if user == nil {
		utils.WriteError(w, http.StatusNotFound, fmt.Errorf("user not found"))
		return
	}

	profilePicture, err := h.userStore.GetUserProfilePicture(user.ID)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	profilePictureURL := ""

	if profilePicture != nil {
		profilePictureURL = profilePicture.Path
	}

	userWithoutPassword := types.UserWithoutPassword{
		ID:          user.ID,
		Name:        user.Name,
		Surname:     user.Surname,
		Email:       user.Email,
		PostalCode:  user.PostalCode,
		UserPicture: profilePictureURL,
		State:       user.State,
		City:        user.City,
		Status:      user.Status,
		RoleID:      user.RoleID,
		CPF:         user.CPF,
		BirthDate:   user.BirthDate,
		Description: user.Description,
		Points:      user.Points,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}

	utils.WriteJSON(w, http.StatusOK, userWithoutPassword)
}

func (h *Handler) HandleGetOwnProfile(w http.ResponseWriter, r *http.Request) {
	userId, found := auth.GetUserIDFromContext(r.Context())

	if !found {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("user not found"))
		return
	}

	h.HandleGetProfile(userId, w, r)
}

func (h *Handler) HandleGetGivenProfile(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	if idStr == "" {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid user ID"))
		return
	}

	id, err := strconv.Atoi(idStr)

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid user ID"))
		return
	}

	h.HandleGetProfile(id, w, r)

}

type NotificationResponse struct {
	ID          int                      `json:"id"`
	IsRead      bool                     `json:"isRead"`
	CreatedAt   time.Time                `json:"createdAt"`
	FromUser    notification.MinimalUser `json:"fromUser"`
	Transaction struct {
		Amount float64 `json:"amount"`
	} `json:"transaction"`
}

func (h *Handler) HandleGetNotifications(w http.ResponseWriter, r *http.Request) {
	userID, found := auth.GetUserIDFromContext(r.Context())
	if !found {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("user not found"))
		return
	}

	notifications, err := h.notificationStore.GetNotificationsByUserID(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to get notifications: %w", err))
		return
	}

	response := make([]NotificationResponse, 0)
	for _, n := range notifications {
		resp := NotificationResponse{
			ID:        n.ID,
			IsRead:    n.IsRead,
			CreatedAt: n.CreatedAt,
			FromUser:  n.FromUser,
		}
		resp.Transaction.Amount = n.Transaction.Amount
		response = append(response, resp)
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) HandleReadNotification(w http.ResponseWriter, r *http.Request) {
	userID, found := auth.GetUserIDFromContext(r.Context())

	if !found {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("user not found"))
		return
	}

	notificationIDStr := r.PathValue("notification_id")

	if notificationIDStr == "" {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid notification ID"))
		return
	}

	notificationID, err := strconv.Atoi(notificationIDStr)

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid notification ID"))
		return
	}

	notification, err := h.notificationStore.ReadNotification(notificationID, userID)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to read notification: %w", err))
		return
	}

	if notification == nil {
		utils.WriteError(w, http.StatusNotFound, fmt.Errorf("notification not found"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, notification)
}

func (h *Handler) HandleTest(w http.ResponseWriter, r *http.Request) {
	userId, found := auth.GetUserIDFromContext(r.Context())

	if !found {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("user not found"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, struct {
		Message string `json:"message"`
	}{
		Message: fmt.Sprintf("ta funfando pae, userId: %d", userId),
	})
}

func (h *Handler) HandleGetUsersByCity(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	userID, found := auth.GetUserIDFromContext(r.Context())

	if !found {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("user not found"))
		return
	}

	if city == "" {
		user, err := h.userStore.GetUserByID(userID)

		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to get user: %w", err))
			return
		}

		city = user.City
	}

	users, err := h.userStore.GetUsersByCity(city)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to get users: %w", err))
		return
	}

	if users == nil {
		users = []*types.User{}
	}

	utils.WriteJSON(w, http.StatusOK, users)
}

func (h *Handler) HandleGetCities(w http.ResponseWriter, r *http.Request) {
	cities, err := h.userStore.GetAllCities()

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to get cities: %w", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, cities)
}

type UpdateDescriptionRequest struct {
	Description string `json:"description" validate:"required"`
}

func (h *Handler) HandleUpdateDescription(w http.ResponseWriter, r *http.Request) {
	var payload UpdateDescriptionRequest

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	userID, found := auth.GetUserIDFromContext(r.Context())

	if !found {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("user not found"))
		return
	}

	user, err := h.userStore.UpdateUserDescription(userID, payload.Description)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to update description: %w", err))
		return
	}

	profilePicture, err := h.userStore.GetUserProfilePicture(user.ID)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to get profile picture: %w", err))
		return
	}

	if profilePicture != nil {
		user.UserPicture = profilePicture.Path
	}

	utils.WriteJSON(w, http.StatusOK, user)
}

func (h *Handler) HandleResendVerificationEmail(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email string `json:"email"`
	}

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// Get user by email
	user, err := h.userStore.GetUserByEmail(payload.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.WriteError(w, http.StatusNotFound, fmt.Errorf("usuário não encontrado"))
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if user.Status == types.StatusActive {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("usuário já está ativo"))
		return
	}

	activationToken, err := auth.CreateJWT([]byte(config.Envs.JWTSecret), user.ID, types.TokenTypeVerify, time.Hour*24)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.mailer.SendConfirmationEmail(user, activationToken)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to send confirmation email: %w", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Email de verificação reenviado com sucesso"})
}

func (h *Handler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("POST /login", h.HandleLogin)
	router.HandleFunc("POST /register", h.HandleRegister)
	router.HandleFunc("POST /refresh-token", auth.HandleTokenRefresh)
	router.HandleFunc("GET /users", auth.WithJWTAuth(h.HandleGetUsersByCity, h.userStore))
	router.HandleFunc("GET /cities", auth.WithJWTAuth(h.HandleGetCities, h.userStore))
	router.HandleFunc("POST /profile-picture", auth.WithJWTAuth(h.HandleAddProfilePicture, h.userStore))
	router.HandleFunc("POST /description", auth.WithJWTAuth(h.HandleUpdateDescription, h.userStore))
	router.HandleFunc("GET /profile/{id}", auth.WithJWTAuth(h.HandleGetGivenProfile, h.userStore))
	router.HandleFunc("GET /profile", auth.WithJWTAuth(h.HandleGetOwnProfile, h.userStore))
	router.HandleFunc("POST /auth", auth.WithJWTAuth(h.HandleTest, h.userStore))
	router.HandleFunc("GET /verify-email", h.HandleVerify)
	router.HandleFunc("GET /notifications", auth.WithJWTAuth(h.HandleGetNotifications, h.userStore))
	router.HandleFunc("POST /notification/{notification_id}/read", auth.WithJWTAuth(h.HandleReadNotification, h.userStore))
	router.HandleFunc("/resend-verification", h.HandleResendVerificationEmail)
}
