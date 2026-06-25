package client

import (
	"context"

	"github.com/nextlevelbuilder/goclaw-cli/wsclient"
)

// FollowHandler is an alias for wsclient.FollowHandler.
// It is called for every event received from the stream.
type FollowHandler = wsclient.FollowHandler

// FollowConfig is an alias for wsclient.FollowConfig.
// It configures FollowStream behaviour.
type FollowConfig = wsclient.FollowConfig

// FollowStream calls method on a fresh WebSocket connection and delivers events
// to handler until ctx is cancelled, handler returns error, or retries exhausted.
//
// This is a forwarding function to wsclient.FollowStream.
func FollowStream(ctx context.Context, serverURL, token, userID string, insecure bool,
	method string, params any, handler FollowHandler, cfg *FollowConfig,
) error {
	return wsclient.FollowStream(ctx, serverURL, token, userID, insecure, method, params, handler, cfg)
}
