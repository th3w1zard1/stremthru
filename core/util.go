package core

import (
	"encoding/json"
	"fmt"
	"strings"
)

func UnmarshalJSON(statusCode int, body []byte, v interface{}) error {
	err := json.Unmarshal(body, v)
	if err == nil {
		return nil
	}

	bodySample := string(body)
	if len(bodySample) > 500 {
		bodySample = bodySample[0:500] + " ..."
	}

	bodySample = strings.Replace(bodySample, "\n", "\\n", -1)

	return fmt.Errorf(
		"Couldn't deserialize JSON (response status: %v, body sample: '%s'): %v",
		statusCode, bodySample, err,
	)
}
