package client

import "encoding/json"

// ConfigGetResult is the response payload for config.get.
type ConfigGetResult struct {
	Config json.RawMessage `json:"config"`
	Hash   string          `json:"hash"`
	Path   string          `json:"path"`
}

// ConfigGet returns the current masked config, its hash, and file path.
func (ws *WSClient) ConfigGet() (*ConfigGetResult, error) {
	payload, err := ws.Call("config.get", nil)
	if err != nil {
		return nil, err
	}
	var result ConfigGetResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ConfigApplyParams are the params for config.apply and config.patch.
type ConfigApplyParams struct {
	Raw      string `json:"raw"`
	BaseHash string `json:"baseHash,omitempty"`
}

// ConfigApplyResult is the response payload for config.apply and config.patch.
type ConfigApplyResult struct {
	OK      bool            `json:"ok"`
	Path    string          `json:"path"`
	Config  json.RawMessage `json:"config"`
	Hash    string          `json:"hash"`
	Restart bool            `json:"restart"`
}

// ConfigApply replaces the entire config with raw JSON5 content.
func (ws *WSClient) ConfigApply(params ConfigApplyParams) (*ConfigApplyResult, error) {
	payload, err := ws.Call("config.apply", params)
	if err != nil {
		return nil, err
	}
	var result ConfigApplyResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ConfigPatch merges a partial JSON5 config update into the current config.
func (ws *WSClient) ConfigPatch(params ConfigApplyParams) (*ConfigApplyResult, error) {
	payload, err := ws.Call("config.patch", params)
	if err != nil {
		return nil, err
	}
	var result ConfigApplyResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ConfigSchemaResult is the response payload for config.schema.
type ConfigSchemaResult struct {
	JSON json.RawMessage `json:"json"`
}

// ConfigSchema returns the config JSON schema for UI form generation.
func (ws *WSClient) ConfigSchema() (*ConfigSchemaResult, error) {
	payload, err := ws.Call("config.schema", nil)
	if err != nil {
		return nil, err
	}
	var result ConfigSchemaResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
