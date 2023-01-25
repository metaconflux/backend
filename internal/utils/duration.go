package utils

import (
	"encoding/json"
	"fmt"
	"time"
)

type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("%s", time.Duration(d)))
}

func (duration *Duration) UnmarshalJSON(b []byte) error {
	var unmarshalledJson interface{}

	err := json.Unmarshal(b, &unmarshalledJson)
	if err != nil {
		return err
	}

	switch value := unmarshalledJson.(type) {
	case float64:
		*duration = Duration(time.Duration(value))
	case string:
		parsed, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*duration = Duration(parsed)
	default:
		return fmt.Errorf("invalid duration: %#v", unmarshalledJson)
	}

	return nil
}

func (d Duration) ToMinute() int64 {
	return int64(time.Duration(d) / time.Minute)
}
