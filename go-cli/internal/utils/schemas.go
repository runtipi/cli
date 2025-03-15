package utils

import (
	"encoding/json"
	"strconv"
)

// StringOrInt represents a value that can be either a string or an integer
type StringOrInt struct {
	value string
}

func (s StringOrInt) AsString() string {
	return s.value
}

func RawToString(value json.RawMessage) string {
	var strValue string
	if err := json.Unmarshal(value, &strValue); err == nil {
		return strValue
	}

	var intValue int
	if err := json.Unmarshal(value, &intValue); err == nil {
		return strconv.Itoa(intValue)
	}

	return ""
}

// SettingsSchema represents the JSON settings file structure
type SettingsSchema struct {
	AppDataPath  string          `json:"appDataPath,omitempty"`
	StoragePath  string          `json:"storagePath,omitempty"`
	InternalIP   string          `json:"internalIp,omitempty"`
	NginxPort    json.RawMessage `json:"nginxPort"`
	NginxSSLPort json.RawMessage `json:"nginxSslPort"`
	PostgresPort json.RawMessage `json:"postgresPort"`
	Domain       string          `json:"domain,omitempty"`
	LocalDomain  string          `json:"localDomain,omitempty"`
}

