package wiretest

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func AssertNotCalledHandler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "expected handler not to be called")
	}
}
