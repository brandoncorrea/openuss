package wiretest

import "net/http"

type BadTransport struct {
	Error error
}

func (t *BadTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, t.Error
}

func NewErrorClient(err error) *http.Client {
	return &http.Client{Transport: &BadTransport{Error: err}}
}
