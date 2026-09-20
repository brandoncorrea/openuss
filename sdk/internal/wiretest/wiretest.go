package wiretest

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type BadTransport struct {
	Error error
}

func (t *BadTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, t.Error
}

func NewErrorClient(err error) *http.Client {
	return &http.Client{Transport: &BadTransport{Error: err}}
}

func AssertNotCalledHandler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "expected handler not to be called")
	}
}
