package scd

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound     = errors.New("scd: operational intent not found")
	ErrRejected     = errors.New("scd: operational intent rejected")
	ErrNotSupported = errors.New("scd: operation not supported")
	ErrConflict     = fmt.Errorf("%w: conflicts with a higher-priority operational intent", ErrRejected)
)
