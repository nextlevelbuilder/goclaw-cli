package client

import "encoding/json"

// AgentInfo describes a single agent as returned by agents.list.
type AgentInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Model     string `json:"model"`
	Provider  string `json:"provider"`
	AgentType string `json:"agentType"`
	Status    string `json:"status"`
	IsRunning bool   `json:"isRunning"`
}

// AgentsListResult is the response payload for agents.list.
type AgentsListResult struct {
	Agents []AgentInfo `json:"agents"`
}

// AgentsList lists agents visible to the caller.
func (ws *WSClient) AgentsList() (*AgentsListResult, error) {
	payload, err := ws.Call("agents.list", nil)
	if err != nil {
		return nil, err
	}
	var result AgentsListResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AgentParams are the params for agent and agent.wait.
type AgentParams struct {
	AgentID string `json:"agentId"`
}

// AgentResult is the response payload for the agent method.
type AgentResult struct {
	ID        string `json:"id"`
	IsRunning bool   `json:"isRunning"`
}

// Agent returns the running state for a single agent.
func (ws *WSClient) Agent(params AgentParams) (*AgentResult, error) {
	payload, err := ws.Call("agent", params)
	if err != nil {
		return nil, err
	}
	var result AgentResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AgentWaitResult is the response payload for agent.wait.
type AgentWaitResult struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// AgentWait waits for (or reports the current status of) an agent.
func (ws *WSClient) AgentWait(params AgentParams) (*AgentWaitResult, error) {
	payload, err := ws.Call("agent.wait", params)
	if err != nil {
		return nil, err
	}
	var result AgentWaitResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
