package template

import (
	"encoding/base64"
	"encoding/json"

	hclog "github.com/hashicorp/go-hclog"
)

// Logger base logger for all template functions
var Logger hclog.Logger

func base64encode(v []byte) string {
	result := base64.StdEncoding.EncodeToString(v)
	Logger.Debug("base64encode", "data", result)

	return result
}

func base64decode(v string) []byte {
	data, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		Logger.Error("base64decode", "error", err)
	}

	return data
}

func toJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		Logger.Error("jsonencode", "error", err)
	}

	return string(data)
}
