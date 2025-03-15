package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"os"
	"runtime"
)

func GetInternalIP() string {
	conn, err := net.Dial("udp", "9.9.9.9:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

// GetArchitecture returns the system architecture
func GetArchitecture() string {
	return runtime.GOARCH
}

// GetSeed reads or creates a seed file
func GetSeed(rootDir string) (string, error) {
	seedPath := rootDir + "/state/seed"
	seed, err := os.ReadFile(seedPath)
	if err != nil {
		return "", err
	}
	return string(seed), nil
}

// DeriveEntropy generates a deterministic string from a key and seed
func DeriveEntropy(key, seed string) string {
	hash := sha256.Sum256([]byte(key + seed))
	return hex.EncodeToString(hash[:])[:32]
}

