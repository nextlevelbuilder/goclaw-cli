package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/nextlevelbuilder/goclaw-cli/internal/client"
)

// newHTTP creates an authenticated HTTP client from current config.
func newHTTP() (*client.HTTPClient, error) {
	if cfg.Server == "" {
		return nil, client.ErrServerRequired
	}
	if cfg.Token == "" {
		return nil, client.ErrNotAuthenticated
	}
	return client.NewHTTPClient(cfg.Server, cfg.Token, cfg.Insecure), nil
}

// newWS creates an authenticated WebSocket client.
func newWS(userID string) (*client.WSClient, error) {
	if cfg.Server == "" {
		return nil, client.ErrServerRequired
	}
	if cfg.Token == "" {
		return nil, client.ErrNotAuthenticated
	}
	if userID == "" {
		userID = "cli"
	}
	return client.NewWSClient(cfg.Server, cfg.Token, userID, cfg.Insecure), nil
}

// unmarshalList is a helper to unmarshal JSON array responses.
func unmarshalList(data json.RawMessage) []map[string]any {
	var list []map[string]any
	_ = json.Unmarshal(data, &list)
	return list
}

// unmarshalMap is a helper to unmarshal JSON object responses.
func unmarshalMap(data json.RawMessage) map[string]any {
	var m map[string]any
	_ = json.Unmarshal(data, &m)
	return m
}

// str safely gets a string from a map.
func str(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

// readContent reads content from flag value: "@file" reads from file, otherwise literal.
func readContent(val string) (string, error) {
	if strings.HasPrefix(val, "@") {
		data, err := os.ReadFile(val[1:])
		if err != nil {
			return "", fmt.Errorf("read file %s: %w", val[1:], err)
		}
		return string(data), nil
	}
	return val, nil
}

// buildBody creates a map from flag values, skipping empty strings.
func buildBody(pairs ...any) map[string]any {
	body := make(map[string]any)
	for i := 0; i < len(pairs)-1; i += 2 {
		key := pairs[i].(string)
		val := pairs[i+1]
		switch v := val.(type) {
		case string:
			if v != "" {
				body[key] = v
			}
		case int:
			if v != 0 {
				body[key] = v
			}
		case bool:
			body[key] = v
		default:
			if v != nil {
				body[key] = v
			}
		}
	}
	return body
}
