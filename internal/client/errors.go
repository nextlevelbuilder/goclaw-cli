package client

import "github.com/nextlevelbuilder/goclaw-cli/wsclient"

// APIError is an alias for wsclient.APIError.
// It represents a structured error from the GoClaw server.
type APIError = wsclient.APIError

// ErrNotAuthenticated is returned when no credentials are configured.
var ErrNotAuthenticated = wsclient.ErrNotAuthenticated

// ErrServerRequired is returned when no server URL is configured.
var ErrServerRequired = wsclient.ErrServerRequired
