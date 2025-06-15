package auth

import (
	"fmt"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	authorization := headers.Get("Authorization")

	if authorization == "" {
		return "", fmt.Errorf("ApiKey not provided")
	}

	key := strings.Replace(authorization, "ApiKey", "", -1)
	key = strings.TrimSpace(key)

	return key, nil
}
