package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/nextlevelbuilder/goclaw-cli/client"
)

// newHTTP creates an authenticated HTTP client from current config.
func newHTTP() (*client.HTTPClient, error) {
	if cfg.Server == "" {
		return nil, client.ErrServerRequired
	}
	if cfg.Token == "" {
		return nil, client.ErrNotAuthenticated
	}
	c := client.NewHTTPClient(cfg.Server, cfg.Token, cfg.Insecure)
	c.TenantID = cfg.TenantID
	return c, nil
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

func decodeRawResponse(resp *http.Response) (map[string]any, error) {
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, apiErrorFromRawBody(resp.StatusCode, data)
	}
	if len(data) == 0 {
		return map[string]any{}, nil
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return out, nil
}

func rawResponseError(resp *http.Response) error {
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	return apiErrorFromRawBody(resp.StatusCode, data)
}

func apiErrorFromRawBody(status int, data []byte) error {
	message := strings.TrimSpace(string(data))
	var env struct {
		Error any `json:"error"`
	}
	if err := json.Unmarshal(data, &env); err == nil {
		switch e := env.Error.(type) {
		case string:
			message = e
		case map[string]any:
			code := fmt.Sprintf("%v", e["code"])
			msg := fmt.Sprintf("%v", e["message"])
			if code != "" && code != "<nil>" && msg != "" && msg != "<nil>" {
				return &client.APIError{StatusCode: status, Code: code, Message: msg, Details: e["details"]}
			}
		}
	}
	if message == "" {
		message = fmt.Sprintf("HTTP %d", status)
	}
	return &client.APIError{StatusCode: status, Code: apiErrorCodeForStatus(status), Message: message}
}

func apiErrorCodeForStatus(status int) string {
	switch status {
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return "INVALID_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "TENANT_ACCESS_REVOKED"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "FAILED_PRECONDITION"
	case http.StatusTooManyRequests:
		return "RESOURCE_EXHAUSTED"
	default:
		if status >= 500 {
			return "INTERNAL"
		}
		return "UNKNOWN"
	}
}
