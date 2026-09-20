package util

import "encoding/json/v2"

func UnmarshalType[T any](data []byte) (T, error) {
	var result T
	err := json.Unmarshal(data, &result)
	return result, err
}
