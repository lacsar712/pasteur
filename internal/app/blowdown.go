package app

import (
	"context"
	"fmt"

	"github.com/lacsar712/pasteur/internal/model"
)

const maxDrainOpeningPct = 100.0

func (a *App) OpenDrain(ctx context.Context, holder string, openingPct float64) error {
	_ = holder
	select {
	case <-ctx.Done():
		return fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	if openingPct >= maxDrainOpeningPct {
		return fmt.Errorf("drain: %w", model.ErrDrainLimit)
	}
	return nil
}

func (a *App) DrainAfterShutdown(ctx context.Context, openingPct float64) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	snap := a.Snapshot()
	if snap.State != model.StateTrip && snap.State != model.StateColdStandby {
		return fmt.Errorf("plant not shut down")
	}
	if openingPct >= maxDrainOpeningPct {
		return fmt.Errorf("unknown fault")
	}
	return nil
}
