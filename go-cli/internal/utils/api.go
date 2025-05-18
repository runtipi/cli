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
	jwt_secret := GetEnvValue("JWT_SECRET")

	if jwt_secret == "" {
		panic("JWT_SECRET is not set")
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
	tokenString, err := token.SignedString([]byte(jwt_secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	return tokenString, nil
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
