package util

import "encoding/json/v2"

func UnmarshalType[T any](body []byte) (T, error) {
	var result T
	err := json.Unmarshal(body, &result)
	return result, err
}
