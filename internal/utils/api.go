// Package utils provides utility functions
package utils

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims
}

func CreateToken() (string, error) {
	jwtSecret := GetEnvValue("JWT_SECRET")

	if jwtSecret == "" {
		return "", fmt.Errorf("JWT_SECRET environment variable is not set. Please set it before running this command")
	}

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "cli",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	return tokenString, nil
}

func GetAPIBaseURL(endpoint string) string {
	envMap := GetEnvMap()

	internalIP := "localhost"
	if ip, ok := envMap["INTERNAL_IP"]; ok && ip != "" {
		internalIP = ip
	}

	nginxPort := DefaultPort
	if port, ok := envMap["NGINX_PORT"]; ok && port != "" {
		nginxPort = port
	}

	return fmt.Sprintf("http://%s:%s/api/%s", internalIP, nginxPort, endpoint)
}

func APIRequest(url string, method string, jsonBody string) (*http.Response, error) {
	client := &http.Client{
		Timeout: time.Second * 30,
	}

	token, err := CreateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %v", err)
	}

	var req *http.Request

	if jsonBody != "" {
		req, err = http.NewRequest(method, url, bytes.NewBuffer([]byte(jsonBody)))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}

	return resp, nil
}
