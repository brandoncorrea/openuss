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
	"time"

	"bwawan.com/openuss/internal/api"
	"bwawan.com/openuss/internal/auth"
)

const DefaultTimeout = 10 * time.Second

type Client struct {
	TokenSource auth.TokenSource
	HTTP        *http.Client
}

func New(tokenSource auth.TokenSource, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: DefaultTimeout}
	}
	return &Client{
		TokenSource: tokenSource,
		HTTP:        httpClient,
	}
}

func (client *Client) Get(ctx context.Context, endpoint string, scopes ...api.RequiredScope) (*http.Response, error) {
	return client.Do(ctx, http.MethodGet, endpoint, nil, scopes...)
}

func (client *Client) Post(ctx context.Context, endpoint string, body any, scopes ...api.RequiredScope) (*http.Response, error) {
	return client.Do(ctx, http.MethodPost, endpoint, body, scopes...)
}

func (client *Client) Do(
	ctx context.Context,
	method string,
	endpoint string,
	body any,
	scopes ...api.RequiredScope,
) (*http.Response, error) {
	requestBody, err := encodeBody(body)
	if err != nil {
		return nil, err
	}
	audience, err := audienceOf(endpoint)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, method, endpoint, requestBody)
	if err != nil {
		return nil, fmt.Errorf("utmclient: failed to create request: %w", err)
	}

	token, err := client.TokenSource.Token(ctx, audience, scopes...)
	if err != nil {
		return nil, fmt.Errorf("utmclient: failed to acquire auth token: %w", err)
	}

	request.Header.Add("Authorization", "Bearer "+token)
	response, err := client.HTTP.Do(request)
	if err != nil {
		return nil, fmt.Errorf("utmclient: failed to make request: %w", err)
	}
	return response, nil
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
