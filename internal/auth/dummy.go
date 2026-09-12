package auth

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"bwawan.com/openuss/internal/api"
	"bwawan.com/openuss/internal/util"
)

type DummyOAuth struct {
	Endpoint *url.URL
	Subject  string
	Client   *http.Client
}

const maxErrorDetail = 512

type DummyTokenResponse struct {
	AccessToken string `json:"access_token"`
}

func NewDummyOAuth(endpoint string, subject string, client *http.Client) (*DummyOAuth, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("auth: parsing token endpoint %q: %w", endpoint, err)
	}
	if !parsed.IsAbs() || parsed.Host == "" {
		return nil, fmt.Errorf("auth: token endpoint %q is not an absolute URL", endpoint)
	}
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return nil, errors.New("auth: subject is required")
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &DummyOAuth{
		Endpoint: parsed,
		Subject:  subject,
		Client:   client,
	}, nil
}

func (auth *DummyOAuth) Token(ctx context.Context, audience string, requiredScopes ...api.RequiredScope) (string, error) {
	audience = strings.TrimSpace(audience)
	if audience == "" {
		return "", errors.New("auth: audience is required")
	}

	scopes := util.Map(requiredScopes, util.TrimString)
	if slices.ContainsFunc(scopes, util.IsBlank) {
		return "", errors.New("auth: scope is empty")
	}

	if len(scopes) == 0 {
		return "", errors.New("auth: at least one scope is required")
	}

	endpoint := *auth.Endpoint
	endpoint.RawQuery = url.Values{
		"grant_type":        {"client_credentials"},
		"scope":             {strings.Join(scopes, " ")},
		"intended_audience": {audience},
		"issuer":            {"dummy"},
		"sub":               {auth.Subject},
	}.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return "", fmt.Errorf("auth: building token request: %w", err)
	}
	response, err := auth.Client.Do(request)
	if err != nil {
		return "", fmt.Errorf("auth: requesting token: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth: token endpoint returned %s: %s", response.Status, responseBody(response.Body))
	}
	var body DummyTokenResponse
	if err := json.UnmarshalRead(response.Body, &body); err != nil {
		return "", fmt.Errorf("auth: decoding token response: %w", err)
	}
	if util.IsBlank(body.AccessToken) {
		return "", errors.New("auth: token response carried no access_token")
	}
	return body.AccessToken, nil
}

func responseBody(body io.Reader) string {
	detail, err := io.ReadAll(io.LimitReader(body, maxErrorDetail))
	if err != nil {
		return "<unreadable body>"
	}
	return string(detail)
}
