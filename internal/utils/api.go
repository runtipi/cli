// Package utils provides utility functions
package utils

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims
}

func CreateToken() (string, error) {
	return CreateTokenWithContext(context.Background())
}

func CreateTokenWithContext(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	secret, err := exec.CommandContext(ctx, "docker", "exec", "runtipi", "cat", "/data/.secrets/cli-secret").Output()
	if ctx.Err() != nil {
		return "", fmt.Errorf("reading CLI authentication key: %w", ctx.Err())
	}
	if err != nil {
		return "", fmt.Errorf("unable to read CLI authentication key through Docker; start the updated backend first: %w", err)
	}

	decoded, err := hex.DecodeString(string(secret))
	if err != nil || len(decoded) != 32 || string(secret) != hex.EncodeToString(decoded) {
		return "", fmt.Errorf("invalid CLI authentication key; restart the backend")
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
	tokenString, err := token.SignedString(secret)
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
	return APIRequestWithContext(context.Background(), url, method, jsonBody)
}

func APIRequestWithContext(ctx context.Context, url string, method string, jsonBody string) (*http.Response, error) {
	client := &http.Client{
		Timeout: time.Second * 30,
	}

	token, err := CreateTokenWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}

	var req *http.Request

	if jsonBody != "" {
		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer([]byte(jsonBody)))
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}
