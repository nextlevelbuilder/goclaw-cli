package client

import "encoding/json"

// CronSchedule describes a cron job's trigger (matches store.CronSchedule).
type CronSchedule struct {
	Kind    string `json:"kind"` // "at", "every", "cron"
	AtMS    *int64 `json:"atMs,omitempty"`
	EveryMS *int64 `json:"everyMs,omitempty"`
	Expr    string `json:"expr,omitempty"`
	TZ      string `json:"tz,omitempty"`
}

// CronCommandSpec is a deterministic shell command run by a cron job (matches
// store.CronCommandSpec).
type CronCommandSpec struct {
	Argv                   []string          `json:"argv,omitempty"`
	Cwd                    string            `json:"cwd,omitempty"`
	Env                    map[string]string `json:"env,omitempty"`
	Input                  string            `json:"input,omitempty"`
	TimeoutSeconds         int               `json:"timeoutSeconds,omitempty"`
	NoOutputTimeoutSeconds int               `json:"noOutputTimeoutSeconds,omitempty"`
	OutputMaxBytes         int               `json:"outputMaxBytes,omitempty"`
}

// CronJobPatch is a partial update for cron.update (matches store.CronJobPatch).
type CronJobPatch struct {
	Name           string           `json:"name,omitempty"`
	AgentID        *string          `json:"agentId,omitempty"`
	Enabled        *bool            `json:"enabled,omitempty"`
	Schedule       *CronSchedule    `json:"schedule,omitempty"`
	Message        string           `json:"message,omitempty"`
	Command        *CronCommandSpec `json:"command,omitempty"`
	DeleteAfterRun *bool            `json:"deleteAfterRun,omitempty"`
	Stateless      *bool            `json:"stateless,omitempty"`
	Deliver        *bool            `json:"deliver,omitempty"`
	DeliverChannel *string          `json:"deliverChannel,omitempty"`
	DeliverTo      *string          `json:"deliverTo,omitempty"`
	WakeHeartbeat  *bool            `json:"wakeHeartbeat,omitempty"`
	ProviderID     *string          `json:"providerId,omitempty"`
	Model          *string          `json:"model,omitempty"`
}

// CronJob describes a cron job as returned by the server (fields loosely
// typed as json.RawMessage where the exact shape isn't load-bearing).
type CronJob struct {
	ID             string          `json:"id"`
	TenantID       string          `json:"tenantId,omitempty"`
	Name           string          `json:"name"`
	AgentID        string          `json:"agentId,omitempty"`
	UserID         string          `json:"userId,omitempty"`
	Enabled        bool            `json:"enabled"`
	Schedule       CronSchedule    `json:"schedule"`
	Payload        json.RawMessage `json:"payload"`
	State          json.RawMessage `json:"state"`
	CreatedAtMS    int64           `json:"createdAtMs"`
	UpdatedAtMS    int64           `json:"updatedAtMs"`
	DeleteAfterRun bool            `json:"deleteAfterRun,omitempty"`
	Stateless      bool            `json:"stateless"`
	Deliver        bool            `json:"deliver"`
	DeliverChannel string          `json:"deliverChannel"`
}

// CronListParams are the params for cron.list.
type CronListParams struct {
	IncludeDisabled bool `json:"includeDisabled,omitempty"`
}

// CronListResult is the response payload for cron.list.
type CronListResult struct {
	Jobs   []CronJob       `json:"jobs"`
	Status json.RawMessage `json:"status"`
}

// CronList lists cron jobs.
func (ws *WSClient) CronList(params CronListParams) (*CronListResult, error) {
	payload, err := ws.Call("cron.list", params)
	if err != nil {
		return nil, err
	}
	var result CronListResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CronCreateParams are the params for cron.create.
type CronCreateParams struct {
	Name           string           `json:"name"`
	Schedule       CronSchedule     `json:"schedule"`
	Message        string           `json:"message,omitempty"`
	Command        *CronCommandSpec `json:"command,omitempty"`
	Deliver        bool             `json:"deliver,omitempty"`
	DeliverChannel string           `json:"deliverChannel,omitempty"`
	DeliverTo      string           `json:"deliverTo,omitempty"`
	WakeHeartbeat  bool             `json:"wakeHeartbeat,omitempty"`
	Stateless      *bool            `json:"stateless,omitempty"`
	AgentID        string           `json:"agentId"`
}

// CronJobResult wraps a single cron job, as returned by cron.create and cron.update.
type CronJobResult struct {
	Job CronJob `json:"job"`
}

// CronCreate creates a new cron job.
func (ws *WSClient) CronCreate(params CronCreateParams) (*CronJobResult, error) {
	payload, err := ws.Call("cron.create", params)
	if err != nil {
		return nil, err
	}
	var result CronJobResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CronUpdateParams are the params for cron.update.
type CronUpdateParams struct {
	JobID string       `json:"jobId"`
	Patch CronJobPatch `json:"patch"`
}

// CronUpdate applies a partial update to a cron job.
func (ws *WSClient) CronUpdate(params CronUpdateParams) (*CronJobResult, error) {
	payload, err := ws.Call("cron.update", params)
	if err != nil {
		return nil, err
	}
	var result CronJobResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CronJobIDParams are the params shared by cron.delete, cron.run and cron.status(single-job variants).
type CronJobIDParams struct {
	JobID string `json:"jobId"`
}

// CronDeleteResult is the response payload for cron.delete.
type CronDeleteResult struct {
	Deleted bool `json:"deleted"`
}

// CronDelete deletes a cron job.
func (ws *WSClient) CronDelete(params CronJobIDParams) (*CronDeleteResult, error) {
	payload, err := ws.Call("cron.delete", params)
	if err != nil {
		return nil, err
	}
	var result CronDeleteResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CronToggleParams are the params for cron.toggle.
type CronToggleParams struct {
	JobID   string `json:"jobId"`
	Enabled bool   `json:"enabled"`
}

// CronToggleResult is the response payload for cron.toggle.
type CronToggleResult struct {
	JobID   string `json:"jobId"`
	Enabled bool   `json:"enabled"`
}

// CronToggle enables or disables a cron job.
func (ws *WSClient) CronToggle(params CronToggleParams) (*CronToggleResult, error) {
	payload, err := ws.Call("cron.toggle", params)
	if err != nil {
		return nil, err
	}
	var result CronToggleResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CronRunParams are the params for cron.run.
type CronRunParams struct {
	JobID string `json:"jobId"`
	Mode  string `json:"mode,omitempty"` // "force" or "due" (default)
}

// CronRunResult is the response payload for cron.run.
type CronRunResult struct {
	OK  bool `json:"ok"`
	Ran bool `json:"ran"`
}

// CronRun triggers an immediate (background) run of a cron job.
func (ws *WSClient) CronRun(params CronRunParams) (*CronRunResult, error) {
	payload, err := ws.Call("cron.run", params)
	if err != nil {
		return nil, err
	}
	var result CronRunResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CronRunsParams are the params for cron.runs.
type CronRunsParams struct {
	JobID  string `json:"jobId,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}

// CronRunsResult is the response payload for cron.runs.
type CronRunsResult struct {
	Entries json.RawMessage `json:"entries"`
	Total   int             `json:"total"`
}

// CronRuns returns the run log entries for a cron job.
func (ws *WSClient) CronRuns(params CronRunsParams) (*CronRunsResult, error) {
	payload, err := ws.Call("cron.runs", params)
	if err != nil {
		return nil, err
	}
	var result CronRunsResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CronStatus returns the scheduler's overall status object (raw, shape not stabilized).
func (ws *WSClient) CronStatus() (json.RawMessage, error) {
	return ws.Call("cron.status", nil)
}
