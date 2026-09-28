package wiretest

import (
	"encoding/json/v2"
	"testing"

	"github.com/stretchr/testify/require"
)

func RequireJSONEncoding(t *testing.T, expected map[string]any, actual any) {
	t.Helper()
	result, err := json.Marshal(actual)
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(result, &out))
	require.Equal(t, expected, out)
}
