package utmclient

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"bwawan.com/openuss/internal/api"
	"bwawan.com/openuss/internal/auth"
	"bwawan.com/openuss/internal/httpclient"
)

type Client struct {
	TokenSource auth.TokenSource
	HTTP        *httpclient.Client
}

func New(tokenSource auth.TokenSource, client *http.Client) *Client {
	return &Client{
		TokenSource: tokenSource,
		HTTP:        httpclient.New(client),
	}
}

func (client *Client) Get(ctx context.Context, endpoint string, scopes ...api.RequiredScope) (httpclient.Response, error) {
	return client.Do(ctx, http.MethodGet, endpoint, nil, scopes...)
}

func (client *Client) Post(ctx context.Context, endpoint string, body any, scopes ...api.RequiredScope) (httpclient.Response, error) {
	return client.Do(ctx, http.MethodPost, endpoint, body, scopes...)
}

func (client *Client) Do(
	ctx context.Context,
	method string,
	endpoint string,
	body any,
	scopes ...api.RequiredScope,
) (httpclient.Response, error) {
	requestBody, err := encodeBody(body)
	if err != nil {
		return httpclient.Response{}, err
	}
	audience, err := audienceOf(endpoint)
	if err != nil {
		return httpclient.Response{}, err
	}
	request, err := http.NewRequestWithContext(ctx, method, endpoint, requestBody)
	if err != nil {
		return httpclient.Response{}, fmt.Errorf("utmclient: failed to create request: %w", err)
	}

	token, err := client.TokenSource.Token(ctx, audience, scopes...)
	if err != nil {
		return httpclient.Response{}, fmt.Errorf("utmclient: failed to acquire auth token: %w", err)
	}

	request.Header.Add("Authorization", "Bearer "+token)
	return client.HTTP.Do(request)
}

func audienceOf(endpoint string) (string, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("utmclient: failed to parse url: %w", err)
	}
	hostname := strings.ToLower(parsed.Hostname())
	if hostname == "" {
		return "", fmt.Errorf("utmclient: url %q has no hostname to use as the token audience", endpoint)
	}
	return hostname, nil
}

func encodeBody(body any) (io.Reader, error) {
	if body == nil {
		return http.NoBody, nil
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("utmclient: failed to encode request body: %w", err)
	}
	return bytes.NewReader(encoded), nil
}
