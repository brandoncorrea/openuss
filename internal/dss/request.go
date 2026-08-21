package dss

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"fmt"
	"net/http"

	"bwawan.com/openuss/internal/api"
)

func (dss *DSS) MakeRequest(
	ctx context.Context,
	method string,
	uri string,
	body any,
	scopes ...api.RequiredScope,
) (*http.Response, error) {
	endpoint := dss.Host + uri
	requestBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("dss: failed to encode request body: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("dss: failed to create request: %w", err)
	}
	normalScopes := normalizeScopes(scopes)
	token, err := dss.TokenSource.Token(ctx, dss.Audience, normalScopes...)
	if err != nil {
		return nil, fmt.Errorf("dss: failed to acquire auth token: %w", err)
	}
	request.Header.Add("Authorization", "Bearer "+token)
	response, err := dss.Client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("dss: failed to make request: %w", err)
	}
	return response, nil
}

func normalizeScopes(scopes []api.RequiredScope) []string {
	normalScopes := []string{}
	for _, scope := range scopes {
		normalScopes = append(normalScopes, string(scope))
	}
	return normalScopes
}
