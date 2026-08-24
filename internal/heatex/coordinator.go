package heatex

import (
	"context"
	"fmt"
	"math"

	"github.com/lacsar712/pasteur/internal/clock"
	"github.com/lacsar712/pasteur/internal/model"
)

type Coordinator struct {
	clk     clock.ProcessClock
	burner  *BurnerController
	airflow *AirflowBalancer
	product    *ProductRegulator
	rinse   *clock.RinseWindow
	ignition *clock.IgnitionDelayWindow
	warmup  *clock.HeatexWarmupWindow
}

func NewCoordinator(clk clock.ProcessClock) *Coordinator {
	return &Coordinator{
		clk:      clk,
		burner:   NewBurnerController(clk),
		airflow:  NewAirflowBalancer(clk),
		product:     NewProductRegulator(clk),
		rinse:    clock.NewRinseWindow(clk),
		ignition: clock.NewIgnitionDelayWindow(clk),
		warmup:   clock.NewHeatexWarmupWindow(clk),
	}
}

func (c *Coordinator) Burner() *BurnerController  { return c.burner }
func (c *Coordinator) Airflow() *AirflowBalancer { return c.airflow }
func (c *Coordinator) Product() *ProductRegulator     { return c.product }

func (c *Coordinator) StartRinse(ctx context.Context, snap model.PlantSnapshot) (model.HeatexReading, error) {
	select {
	case <-ctx.Done():
		return snap.Heatex, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	out := snap.Heatex
	out.BurnerPhase = model.BurnerRinse
	out.RinseStartedAt = c.clk.Now()
	out.ProductFlowTPH = 0
	out.AirflowTPH = c.airflow.RinseRate()
	return out, nil
}

func (c *Coordinator) CompleteRinse(snap model.HeatexReading) error {
	return c.rinse.Require(snap.RinseStartedAt)
}

func (c *Coordinator) Ignite(ctx context.Context, snap model.PlantSnapshot) (model.HeatexReading, error) {
	select {
	case <-ctx.Done():
		return snap.Heatex, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	if err := c.rinse.Require(snap.Heatex.RinseStartedAt); err != nil {
		return snap.Heatex, err
	}
	out := snap.Heatex
	out.BurnerPhase = model.BurnerIgnition
	out.IgnitionAt = c.clk.Now()
	out.ProductFlowTPH = c.product.IgnitionRate(snap.Settings)
	out.AirflowTPH = c.airflow.IgnitionRate(snap.Settings)
	out.BeltframeTempF = 400
	return out, nil
}

func (c *Coordinator) Stabilize(snap model.PlantSnapshot) (model.HeatexReading, error) {
	if err := c.ignition.Require(snap.Heatex.IgnitionAt); err != nil {
		return snap.Heatex, err
	}
	out := snap.Heatex
	out.BurnerPhase = model.BurnerStable
	out.ProductFlowTPH = snap.Settings.ProductFlowTPH * 0.5
	out.AirflowTPH = c.airflow.Compute(snap)
	out.ExcessO2Pct = c.airflow.ExcessO2(out)
	out.BeltframeTempF = c.burner.EstimateBeltframeTemp(out)
	return out, nil
}

func (c *Coordinator) RampToLoad(snap model.PlantSnapshot, loadPct float64) model.HeatexReading {
	out := snap.Heatex
	out.ProductFlowTPH = snap.Settings.ProductFlowTPH * loadPct
	out.AirflowTPH = c.airflow.Compute(snap)
	out.ExcessO2Pct = c.airflow.ExcessO2(out)
	out.BeltframeTempF = c.burner.EstimateBeltframeTemp(out)
	return out
}

func (c *Coordinator) Trip(snap model.HeatexReading) model.HeatexReading {
	out := snap
	out.BurnerPhase = model.BurnerTrip
	out.ProductFlowTPH = 0
	out.BeltframeTempF = math.Max(200, out.BeltframeTempF*0.5)
	return out
}

func (c *Coordinator) WarmupReady(snap model.HeatexReading) bool {
	return c.warmup.Ready(snap.IgnitionAt)
}
