package utils

//make a write json utils for http handlers
import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

var Validate = validator.New()

func ParseJSON(r *http.Request, payload any) error {
	if r.Body == nil {
		return fmt.Errorf("missing request body")
	}

	return json.NewDecoder(r.Body).Decode(payload)
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, status int, err error) {
	WriteJSON(w, status, map[string]string{
		"error": err.Error(),
	})
}

func GetInt(v interface{}) (int, error) {
	switch v := v.(type) {
	case float64:
		return int(v), nil
	case string:
		c, err := strconv.Atoi(v)
		if err != nil {
			return 0, err
		}
		return c, nil
	case json.Number:
		c, err := v.Int64()
		if err != nil {
			return 0, err
		}
		return int(c), nil
	default:
		fmt.Println(v)
		return 0, fmt.Errorf("conversion to int from %T not supported", v)
	}
}

func ParseDate(date string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.999Z",
	}

	var t time.Time
	var err error

	for _, f := range formats {
		t, err = time.Parse(f, date)
		if err == nil {
			break
		}
	}

	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format")
	}

	return t, nil

}

func GetHttpClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
	}
}

// https://www.mercadopago.com.br/developers/pt/docs/your-integrations/notifications/webhooks#editor_8
func WebhookHeaderValidator(r *http.Request, secret string) bool {
	// Obtain the x-signature value from the header
	xSignature := r.Header.Get("x-signature")
	xRequestId := r.Header.Get("x-request-id")

	if xSignature == "" || xRequestId == "" {
		// If x-signature or x-request-id is missing, return an error
		return false
	}

	// Obtain Query params related to the request URL
	queryParams := r.URL.Query()

	// Extract the "data.id" from the query params
	dataID := queryParams.Get("data.id")

	// Separating the x-signature into parts
	parts := strings.Split(xSignature, ",")

	// Initializing variables to store ts and hash
	var ts, hash string

	// Iterate over the values to obtain ts and v1
	for _, part := range parts {
		// Split each part into key and value
		keyValue := strings.SplitN(part, "=", 2)
		if len(keyValue) == 2 {
			key := strings.TrimSpace(keyValue[0])
			value := strings.TrimSpace(keyValue[1])
			if key == "ts" {
				ts = value
			} else if key == "v1" {
				hash = value
			}
		}
	}

	// Generate the manifest string
	manifest := fmt.Sprintf("id:%v;request-id:%v;ts:%v;", dataID, xRequestId, ts)

	// Create an HMAC signature defining the hash type and the key as a byte array
	hmac := hmac.New(sha256.New, []byte(secret))
	hmac.Write([]byte(manifest))

	// Obtain the hash result as a hexadecimal string
	sha := hex.EncodeToString(hmac.Sum(nil))

	return sha == hash
}
