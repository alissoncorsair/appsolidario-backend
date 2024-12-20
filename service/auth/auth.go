package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alissoncorsair/appsolidario-backend/config"
	"github.com/alissoncorsair/appsolidario-backend/service/token"
	"github.com/alissoncorsair/appsolidario-backend/types"
	"github.com/alissoncorsair/appsolidario-backend/utils"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type contextKey string

const UserKey contextKey = "userID"

func StoreToken(store token.Store, token *types.Token) error {
	_, err := store.CreateToken(token)
	return err
}

func CreateJWT(secret []byte, userID int, tokenType types.TokenType, expiration time.Duration) (string, error) {
	expirationTime := time.Now().Add(expiration).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID":    strconv.Itoa(userID),
		"exp":       expirationTime,
		"tokenType": tokenType,
	})

	tokenString, err := token.SignedString(secret)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func HandleTokenRefresh(w http.ResponseWriter, r *http.Request) {
	var request struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := utils.ParseJSON(r, &request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid request payload"))
		return
	}

	token, err := ValidateToken(request.RefreshToken, types.TokenTypeRefresh)
	if err != nil || !token.Valid {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid refresh token"))
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token claims"))
		return
	}

	userIDStr := claims["userID"].(string)
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid user ID"))
		return
	}

	// Generate new access token and refresh token
	accessToken, err := CreateJWT([]byte(config.Envs.JWTSecret), userID, types.TokenTypeAccess, time.Duration(config.Envs.JWTExpirationInSeconds)*time.Second)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("could not generate access token"))
		return
	}

	refreshToken, err := CreateJWT([]byte(config.Envs.JWTSecret), userID, types.TokenTypeRefresh, time.Duration(config.Envs.JWTRefreshExpirationInSeconds)*time.Second)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("could not generate refresh token"))
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

func WithJWTAuth(handlerFunc http.HandlerFunc, store types.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := getTokenFromRequest(r)
		//token is validating both access and refresh tokens, but it should only validate access tokens
		token, err := ValidateToken(tokenString, types.TokenTypeAccess)

		if err != nil {
			log.Printf("failed to validate token: %v", err)
			permissionDenied(w)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)

		if !ok {
			log.Printf("failed to validate token: %v", err)
			permissionDenied(w)
			return
		}

		str := claims["userID"].(string)
		userID, err := strconv.Atoi(str)

		if err != nil {
			log.Printf("failed to convert userID to int: %v", err)
			permissionDenied(w)
			return
		}

		u, err := store.GetUserByID(userID)

		if err != nil {
			log.Printf("failed to find user: %v", err)
			permissionDenied(w)
			return
		}

		if u == nil {
			permissionDenied(w)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, UserKey, u.ID)

		r = r.WithContext(ctx)

		handlerFunc(w, r)
	}
}

func getTokenFromRequest(r *http.Request) string {
	tokenAuth := r.Header.Get("Authorization")
	prefix := "Bearer "

	tokenAuth = strings.TrimPrefix(tokenAuth, prefix)

	return tokenAuth
}

func ValidateToken(tokenString string, expectedType types.TokenType) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Envs.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["tokenType"].(string) != string(expectedType) {
		return nil, fmt.Errorf("invalid token type")
	}

	return token, nil
}

func permissionDenied(w http.ResponseWriter) {
	utils.WriteError(w, http.StatusForbidden, fmt.Errorf("permission denied"))
}

func GetUserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(UserKey).(int)
	if !ok {
		return -1, false
	}

	return userID, true
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func ComparePassword(hashedPassword string, plain []byte) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), plain)
	return err == nil
}
