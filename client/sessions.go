package client

import "encoding/json"

// SessionsListParams are the params for sessions.list.
type SessionsListParams struct {
	AgentID string `json:"agentId,omitempty"`
	Channel string `json:"channel,omitempty"`
	Limit   int    `json:"limit,omitempty"`
	Offset  int    `json:"offset,omitempty"`
}

// SessionsListResult is the response payload for sessions.list.
type SessionsListResult struct {
	Sessions json.RawMessage `json:"sessions"`
	Total    int             `json:"total"`
	Limit    int             `json:"limit"`
	Offset   int             `json:"offset"`
}

// SessionsList lists sessions with optional filtering.
func (ws *WSClient) SessionsList(params SessionsListParams) (*SessionsListResult, error) {
	payload, err := ws.Call("sessions.list", params)
	if err != nil {
		return nil, err
	}
	var result SessionsListResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SessionsPreviewParams are the params for sessions.preview.
type SessionsPreviewParams struct {
	Key string `json:"key"`
}

// SessionsPreviewResult is the response payload for sessions.preview.
type SessionsPreviewResult struct {
	Key      string          `json:"key"`
	Messages json.RawMessage `json:"messages"`
	Summary  json.RawMessage `json:"summary,omitempty"`
}

// SessionsPreview returns the message history and summary for a session.
func (ws *WSClient) SessionsPreview(params SessionsPreviewParams) (*SessionsPreviewResult, error) {
	payload, err := ws.Call("sessions.preview", params)
	if err != nil {
		return nil, err
	}
	var result SessionsPreviewResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SessionsPatchParams are the params for sessions.patch.
type SessionsPatchParams struct {
	Key      string            `json:"key"`
	Label    *string           `json:"label,omitempty"`
	Model    *string           `json:"model,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// SessionsPatchResult is the response payload for sessions.patch.
type SessionsPatchResult struct {
	OK  bool   `json:"ok"`
	Key string `json:"key"`
}

// SessionsPatch updates label/model/metadata on a session.
func (ws *WSClient) SessionsPatch(params SessionsPatchParams) (*SessionsPatchResult, error) {
	payload, err := ws.Call("sessions.patch", params)
	if err != nil {
		return nil, err
	}
	var result SessionsPatchResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SessionsKeyParams are the params shared by sessions.delete and sessions.reset.
type SessionsKeyParams struct {
	Key string `json:"key"`
}

// SessionsOKResult is a generic {ok:true} response.
type SessionsOKResult struct {
	OK bool `json:"ok"`
}

// SessionsDelete deletes a session.
func (ws *WSClient) SessionsDelete(params SessionsKeyParams) (*SessionsOKResult, error) {
	payload, err := ws.Call("sessions.delete", params)
	if err != nil {
		return nil, err
	}
	var result SessionsOKResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SessionsReset resets a session's transcript.
func (ws *WSClient) SessionsReset(params SessionsKeyParams) (*SessionsOKResult, error) {
	payload, err := ws.Call("sessions.reset", params)
	if err != nil {
		return nil, err
	}
	var result SessionsOKResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SessionsCompactParams are the params for sessions.compact.
type SessionsCompactParams struct {
	Key      string `json:"key"`
	KeepLast int    `json:"keepLast,omitempty"`
}

// SessionsCompactResult is the response payload for sessions.compact.
type SessionsCompactResult struct {
	OK       bool   `json:"ok"`
	Original int    `json:"original,omitempty"`
	Kept     int    `json:"kept"`
	Message  string `json:"message,omitempty"`
}

// SessionsCompact compacts a session's history, keeping only the most recent messages.
func (ws *WSClient) SessionsCompact(params SessionsCompactParams) (*SessionsCompactResult, error) {
	payload, err := ws.Call("sessions.compact", params)
	if err != nil {
		return nil, err
	}
	var result SessionsCompactResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
