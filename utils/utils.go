package utils

//make a write json utils for http handlers
import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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
