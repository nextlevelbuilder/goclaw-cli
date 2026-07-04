package client

import "encoding/json"

// SkillInfo describes a skill as returned by skills.list. Field names are
// snake_case, matching the server's skills.list response (unlike most other
// methods, which use camelCase).
type SkillInfo struct {
	ID            string   `json:"id,omitempty"`
	Name          string   `json:"name"`
	Slug          string   `json:"slug"`
	Description   string   `json:"description"`
	Source        string   `json:"source"`
	Version       string   `json:"version"`
	IsSystem      bool     `json:"is_system"`
	Enabled       bool     `json:"enabled"`
	Visibility    string   `json:"visibility,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	Status        string   `json:"status,omitempty"`
	Author        string   `json:"author,omitempty"`
	CreatorAgent  *string  `json:"creator_agent,omitempty"`
	ManagerAgents []string `json:"manager_agents,omitempty"`
	MissingDeps   []string `json:"missing_deps,omitempty"`
	TenantEnabled *bool    `json:"tenant_enabled,omitempty"`
}

// SkillsListResult is the response payload for skills.list.
type SkillsListResult struct {
	Skills []SkillInfo `json:"skills"`
}

// SkillsList lists skills visible to the caller.
func (ws *WSClient) SkillsList() (*SkillsListResult, error) {
	payload, err := ws.Call("skills.list", nil)
	if err != nil {
		return nil, err
	}
	var result SkillsListResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SkillsGetParams are the params for skills.get.
type SkillsGetParams struct {
	Name string `json:"name"`
}

// SkillsGetResult is the response payload for skills.get. Embeds SkillInfo's
// fields (snake_case) plus the skill's raw content.
type SkillsGetResult struct {
	SkillInfo
	Content string `json:"content"`
}

// SkillsGet fetches a skill's metadata and content.
func (ws *WSClient) SkillsGet(params SkillsGetParams) (*SkillsGetResult, error) {
	payload, err := ws.Call("skills.get", params)
	if err != nil {
		return nil, err
	}
	var result SkillsGetResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SkillsUpdateParams are the params for skills.update.
type SkillsUpdateParams struct {
	Name    string         `json:"name,omitempty"`
	ID      string         `json:"id,omitempty"`
	Updates map[string]any `json:"updates"`
}

// SkillsUpdateResult is the response payload for skills.update. Note: the
// server returns the string literal "true"/"false" for "ok", not a JSON bool.
type SkillsUpdateResult struct {
	OK string `json:"ok"`
}

// SkillsUpdate updates a skill's metadata by name or id.
func (ws *WSClient) SkillsUpdate(params SkillsUpdateParams) (*SkillsUpdateResult, error) {
	payload, err := ws.Call("skills.update", params)
	if err != nil {
		return nil, err
	}
	var result SkillsUpdateResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
