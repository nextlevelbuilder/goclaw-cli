package client

import "github.com/nextlevelbuilder/goclaw-cli/wsclient"

// WSClient is an alias for wsclient.WSClient.
// It is a WebSocket RPC client implementing GoClaw protocol v3.
type WSClient = wsclient.WSClient

// WSRequest is an alias for wsclient.WSRequest.
// It is a v3 protocol request frame.
type WSRequest = wsclient.WSRequest

// WSResponse is an alias for wsclient.WSResponse.
// It is a v3 protocol response frame.
type WSResponse = wsclient.WSResponse

// WSEvent is an alias for wsclient.WSEvent.
// It is a v3 protocol server push event.
type WSEvent = wsclient.WSEvent

// NewWSClient creates a WebSocket client.
func NewWSClient(serverURL, token, userID string, insecure bool) *WSClient {
	return wsclient.NewWSClient(serverURL, token, userID, insecure)
}
