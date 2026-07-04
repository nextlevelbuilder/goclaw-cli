package client

import "encoding/json"

// ChatMediaItem represents a media file attached to a chat message.
type ChatMediaItem struct {
	Path     string `json:"path"`
	Filename string `json:"filename,omitempty"`
}

// ChatSendParams are the params for chat.send.
type ChatSendParams struct {
	Message    string          `json:"message"`
	AgentID    string          `json:"agentId"`
	SessionKey string          `json:"sessionKey,omitempty"`
	Stream     bool            `json:"stream,omitempty"`
	Media      json.RawMessage `json:"media,omitempty"` // []string (legacy) or []ChatMediaItem
}

// ChatSendResult is the response payload for chat.send.
type ChatSendResult struct {
	RunID     string          `json:"runId"`
	Content   string          `json:"content"`
	Usage     json.RawMessage `json:"usage,omitempty"`
	Thinking  string          `json:"thinking,omitempty"`
	Media     json.RawMessage `json:"media,omitempty"`
	Cancelled bool            `json:"cancelled,omitempty"`
	Injected  bool            `json:"injected,omitempty"`
}

// ChatSend sends a chat message to an agent.
func (ws *WSClient) ChatSend(params ChatSendParams) (*ChatSendResult, error) {
	payload, err := ws.Call("chat.send", params)
	if err != nil {
		return nil, err
	}
	var result ChatSendResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ChatSendStream sends a chat message and streams events until run completion.
func (ws *WSClient) ChatSendStream(params ChatSendParams, onEvent func(*WSEvent)) (*ChatSendResult, error) {
	payload, err := ws.Stream("chat.send", params, onEvent)
	if err != nil {
		return nil, err
	}
	var result ChatSendResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ChatHistoryParams are the params for chat.history.
type ChatHistoryParams struct {
	AgentID    string `json:"agentId"`
	SessionKey string `json:"sessionKey"`
}

// ChatHistoryResult is the response payload for chat.history.
type ChatHistoryResult struct {
	Messages json.RawMessage `json:"messages"`
}

// ChatHistory fetches the message history for a session.
func (ws *WSClient) ChatHistory(params ChatHistoryParams) (*ChatHistoryResult, error) {
	payload, err := ws.Call("chat.history", params)
	if err != nil {
		return nil, err
	}
	var result ChatHistoryResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ChatInjectParams are the params for chat.inject.
type ChatInjectParams struct {
	SessionKey string `json:"sessionKey"`
	Message    string `json:"message"`
	Label      string `json:"label,omitempty"`
}

// ChatInjectResult is the response payload for chat.inject.
type ChatInjectResult struct {
	OK        bool   `json:"ok"`
	MessageID string `json:"messageId"`
}

// ChatInject injects a message into a session transcript without running the agent.
func (ws *WSClient) ChatInject(params ChatInjectParams) (*ChatInjectResult, error) {
	payload, err := ws.Call("chat.inject", params)
	if err != nil {
		return nil, err
	}
	var result ChatInjectResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ChatAbortParams are the params for chat.abort.
type ChatAbortParams struct {
	RunID      string `json:"runId,omitempty"`
	SessionKey string `json:"sessionKey,omitempty"`
}

// ChatAbortResult is the response payload for chat.abort.
type ChatAbortResult struct {
	OK              bool     `json:"ok"`
	Aborted         bool     `json:"aborted"`
	Stopped         bool     `json:"stopped"`
	Forced          bool     `json:"forced"`
	AlreadyAborting bool     `json:"alreadyAborting"`
	NotFound        bool     `json:"notFound"`
	Unauthorized    bool     `json:"unauthorized"`
	RunIDs          []string `json:"runIds"`
}

// ChatAbort cancels running agent invocations for a session or specific run.
func (ws *WSClient) ChatAbort(params ChatAbortParams) (*ChatAbortResult, error) {
	payload, err := ws.Call("chat.abort", params)
	if err != nil {
		return nil, err
	}
	var result ChatAbortResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ChatSessionStatusParams are the params for chat.session.status.
type ChatSessionStatusParams struct {
	SessionKey string `json:"sessionKey"`
}

// ChatActivity describes the current in-flight agent activity for a session.
type ChatActivity struct {
	Phase     string `json:"phase"`
	Tool      string `json:"tool"`
	Iteration int    `json:"iteration"`
}

// ChatSessionStatusResult is the response payload for chat.session.status.
type ChatSessionStatusResult struct {
	IsRunning bool          `json:"isRunning"`
	RunID     string        `json:"runId"`
	Activity  *ChatActivity `json:"activity"`
}

// ChatSessionStatus returns the running state and activity for a session.
func (ws *WSClient) ChatSessionStatus(params ChatSessionStatusParams) (*ChatSessionStatusResult, error) {
	payload, err := ws.Call("chat.session.status", params)
	if err != nil {
		return nil, err
	}
	var result ChatSessionStatusResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
