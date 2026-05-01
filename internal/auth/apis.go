package auth

import (
	"fmt"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	value := headers.Get("Authorization")
	if value == "" {
		return "", fmt.Errorf("No Authentication header")
	}
	after, _ := strings.CutPrefix(value, "ApiKey ")
	clean := strings.TrimSpace(after)
	return clean, nil
}
