package render

import (
	"encoding/json"

	"github.com/davisbuilds/envdiff/internal/model"
)

func JSON(envelope model.JsonEnvelope) (string, error) {
	payload, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return "", err
	}
	return string(payload), nil
}
