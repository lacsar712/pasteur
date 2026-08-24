package heatex

import (
	"math"

	"github.com/lacsar712/pasteur/internal/clock"
	"github.com/lacsar712/pasteur/internal/model"
)

type ProductRegulator struct {
	clk clock.ProcessClock
}

func NewProductRegulator(clk clock.ProcessClock) *ProductRegulator {
	return &ProductRegulator{clk: clk}
}

func (f *ProductRegulator) IgnitionRate(settings model.PlantSettings) float64 {
	return settings.ProductFlowTPH * 0.08
}

func (f *ProductRegulator) ComputeForLoad(settings model.PlantSettings, loadPct float64) float64 {
	loadPct = math.Max(0, math.Min(1, loadPct))
	return settings.ProductFlowTPH * loadPct
}

func (f *ProductRegulator) Ramp(current, target, maxStep float64) float64 {
	delta := target - current
	if math.Abs(delta) <= maxStep {
		return target
	}
	if delta > 0 {
		return current + maxStep
	}
	return current - maxStep
}

func (f *ProductRegulator) BtuPerHour(flowTPH float64) float64 {
	return flowTPH * 19_500_000
}

func (f *ProductRegulator) HeatInputMW(flowTPH float64) float64 {
	return flowTPH * 11.6
}

func (f *ProductRegulator) ValidatePermissive(settings model.PlantSettings, carrierOK, rinseOK bool) error {
	if !rinseOK {
		return model.ErrRinseIncomplete
	}
	if !carrierOK {
		return model.ErrCarrierLevelTrip
	}
	if settings.ProductFlowTPH <= 0 {
		return model.ErrProductPermissive
	}
	return nil
}

func (f *ProductRegulator) MinFlow(settings model.PlantSettings) float64 {
	return settings.ProductFlowTPH * 0.2
}

func (f *ProductRegulator) MaxFlow(settings model.PlantSettings) float64 {
	return settings.ProductFlowTPH * 1.1
}
