package util

import (
	"encoding/json/v2"
	"io"
)

func UnmarshalType[T any](reader io.Reader) (T, error) {
	var result T
	err := json.UnmarshalRead(reader, &result)
	return result, err
}
