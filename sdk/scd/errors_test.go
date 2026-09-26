package scd_test

import (
	"testing"

	"bwawan.com/openuss/sdk/scd"
	"github.com/stretchr/testify/require"
)

func TestErrNotFound(t *testing.T) {
	require.EqualError(t, scd.ErrNotFound, "scd: operational intent not found")
}

func TestErrRejected(t *testing.T) {
	require.EqualError(t, scd.ErrRejected, "scd: operational intent rejected")
}

func TestErrConflict(t *testing.T) {
	message := scd.ErrRejected.Error() + ": conflicts with a higher-priority operational intent"
	require.EqualError(t, scd.ErrConflict, message)
	require.ErrorIs(t, scd.ErrConflict, scd.ErrRejected)
}

func TestErrNotSupported(t *testing.T) {
	require.EqualError(t, scd.ErrNotSupported, "scd: operation not supported")
}
