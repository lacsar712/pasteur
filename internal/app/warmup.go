package app

import (
	"context"
	"fmt"

	"github.com/lacsar712/pasteur/internal/model"
)

func (a *App) WarmupStatus() (ready bool, detail string) {
	snap := a.Snapshot()
	if snap.Heatex.RinseStartedAt.IsZero() {
		return false, "rinse not started"
	}
	if !a.rinseWindow.Ready(snap.Heatex.RinseStartedAt) {
		return false, "rinse window open"
	}
	if !snap.Heatex.IgnitionAt.IsZero() && !a.warmupWindow.Ready(snap.Heatex.IgnitionAt) {
		return false, "heatex warmup window open"
	}
	if !snap.Carrier.LastSwellAt.IsZero() {
		if err := a.carrier.RequireSettled(snap.Carrier); err != nil {
			return false, "carrier swell settling"
		}
	}
	return true, "ready"
}

func (a *App) WaitWarmup(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("%w", model.ErrContextDone)
		default:
		}
		ready, _ := a.WarmupStatus()
		if ready {
			return nil
		}
	}
}

func (a *App) RinseRemaining() string {
	snap := a.Snapshot()
	if snap.Heatex.RinseStartedAt.IsZero() {
		return "not started"
	}
	if a.rinseWindow.Ready(snap.Heatex.RinseStartedAt) {
		return "complete"
	}
	return "in progress"
}

func (a *App) HeatexWarmupRemaining() string {
	snap := a.Snapshot()
	if snap.Heatex.IgnitionAt.IsZero() {
		return "not ignited"
	}
	if a.warmupWindow.Ready(snap.Heatex.IgnitionAt) {
		return "complete"
	}
	return "in progress"
}
