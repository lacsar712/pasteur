package app

import (
	"context"
	"fmt"
	"time"

	"github.com/lacsar712/pasteur/internal/clock"
	"github.com/lacsar712/pasteur/internal/model"
)

func (a *App) advanceClock(d time.Duration) {
	if mc, ok := a.clk.(*clock.ManualClock); ok {
		mc.Advance(d)
		time.Sleep(time.Millisecond)
	} else {
		time.Sleep(d)
	}
}

func (a *App) bindProductLoop(holder string, ctx context.Context) context.Context {
	a.mu.Lock()
	if cancel, ok := a.productLoopCancels[holder]; ok {
		cancel()
	}
	child, cancel := context.WithCancel(ctx)
	a.productLoopCancels[holder] = cancel
	a.mu.Unlock()
	return child
}

func (a *App) cancelProductLoop(holder string) {
	a.mu.Lock()
	if cancel, ok := a.productLoopCancels[holder]; ok {
		cancel()
		delete(a.productLoopCancels, holder)
	}
	a.mu.Unlock()
}

func (a *App) cancelAllProductLoops() {
	a.mu.Lock()
	for holder, cancel := range a.productLoopCancels {
		cancel()
		delete(a.productLoopCancels, holder)
	}
	a.mu.Unlock()
}

func (a *App) CoalFeedTPH() float64 {
	return a.Snapshot().Heatex.ProductFlowTPH
}

func (a *App) RunProductRamp(ctx context.Context, holder string, targetTPH float64) error {
	loopCtx := a.bindProductLoop(holder, ctx)
	defer a.cancelProductLoop(holder)
	for {
		if err := loopCtx.Err(); err != nil {
			return fmt.Errorf("%w", model.ErrContextDone)
		}
		snap := a.Snapshot()
		current := snap.Heatex.ProductFlowTPH
		if current >= targetTPH {
			return nil
		}
		comb := snap.Heatex
		comb.ProductFlowTPH = current + 1.0
		_ = a.store.UpdateHeatex(a.cfg.UnitID, comb)
		a.telemetry.RecordCoalFeed(comb.ProductFlowTPH)
		a.advanceClock(100 * time.Millisecond)
	}
}

func (a *App) RunCoalFeed(ctx context.Context, holder string, steps int) error {
	loopCtx := a.bindProductLoop(holder, ctx)
	defer a.cancelProductLoop(holder)
	for i := 0; steps <= 0 || i < steps; i++ {
		if err := loopCtx.Err(); err != nil {
			return fmt.Errorf("%w", model.ErrContextDone)
		}
		snap := a.Snapshot()
		comb := snap.Heatex
		comb.ProductFlowTPH += 0.5
		_ = a.store.UpdateHeatex(a.cfg.UnitID, comb)
		a.telemetry.RecordCoalFeed(comb.ProductFlowTPH)
		a.advanceClock(100 * time.Millisecond)
	}
	return nil
}
