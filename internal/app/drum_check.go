package app

import (
	"fmt"

	"github.com/lacsar712/pasteur/internal/model"
)

func (a *App) CheckCarrierLevel(snap model.PlantSnapshot) error {
	if snap.Carrier.LevelPercent < model.MinCarrierLevelPercent {
		return fmt.Errorf("%w", model.ErrCarrierLevelLow)
	}
	return nil
}
