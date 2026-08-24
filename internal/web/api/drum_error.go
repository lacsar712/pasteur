package api

import (
	"errors"

	"github.com/lacsar712/pasteur/internal/model"
)

func classifyCarrierError(err error) (string, bool) {
	if errors.Is(err, model.ErrCarrierLevelLow) {
		return "carrier_level_low", true
	}
	return "", false
}
