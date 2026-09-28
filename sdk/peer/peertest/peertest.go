package peertest

import (
	"net/http"
	"testing"

	"bwawan.com/openuss/sdk/peer"
	"bwawan.com/openuss/sdk/utmclient/utmclienttest"
)

func NewPeer(t *testing.T, handler http.HandlerFunc) *peer.UTMClient {
	t.Helper()
	return peer.New(utmclienttest.NewClient(t, handler))
}
