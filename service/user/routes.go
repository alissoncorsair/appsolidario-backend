package user

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/alissoncorsair/appsolidario-backend/config"
	"github.com/alissoncorsair/appsolidario-backend/service/auth"
	"github.com/alissoncorsair/appsolidario-backend/types"
	"github.com/alissoncorsair/appsolidario-backend/utils"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	store *Store
}

func NewHandler(store *Store) *Handler {
	return &Handler{
		store: store,
	}
}

func (h *Handler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("POST /login", h.HandleLogin)
	router.HandleFunc("POST /register", h.HandleRegister)
	router.HandleFunc("POST /refresh-token", auth.HandleTokenRefresh)
	router.HandleFunc("POST /auth", auth.WithJWTAuth(h.HandleTest, h.store))
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

	_, err := h.store.GetUserByEmail(payload.Email)

	if err == nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user with email %s already exists", payload.Email))
		return
	}

	hashedPass, err := auth.HashPassword(payload.Password)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	var birthDate time.Time
	birthDate, err = time.Parse("2006-01-02", payload.BirthDate)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("invalid birth date"))
		return
	}

	var cpf string
	re := regexp.MustCompile("[^0-9]")
	cpf = re.ReplaceAllString(payload.CPF, "")

	user, err := h.store.CreateUser(&types.User{
		Name:       payload.Name,
		Surname:    payload.Surname,
		Password:   hashedPass,
		Email:      payload.Email,
		PostalCode: payload.PostalCode,
		State:      payload.State,
		City:       payload.City,
		Status:     types.StatusActive,
		RoleID:     types.UserRole(payload.RoleID),
		CPF:        cpf,
		BirthDate:  birthDate,
	})

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, user)
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

	user, err := h.store.GetUserByCPF(payload.CPF)

	if err != nil {
		//should not tell if the user exists or not
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid cpf/password"))
		return
	}
	fmt.Println(user.Password)
	if !auth.ComparePassword(user.Password, []byte(payload.Password)) {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid cpf/password"))
		return
	}

	accessToken, err := auth.CreateJWT([]byte(config.Envs.JWTSecret), user.ID)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	refreshToken, err := auth.CreateRefreshToken([]byte(config.Envs.JWTSecret), user.ID)

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

func (h *Handler) HandleTest(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, struct {
		Message string `json:"message"`
	}{
		Message: "ta funfando pae",
	})
}
