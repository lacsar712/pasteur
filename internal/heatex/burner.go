package heatex

import (
	"math"

	"github.com/lacsar712/pasteur/internal/clock"
	"github.com/lacsar712/pasteur/internal/model"
)

type BurnerController struct {
	clk clock.ProcessClock
}

func NewBurnerController(clk clock.ProcessClock) *BurnerController {
	return &BurnerController{clk: clk}
}

func (b *BurnerController) EstimateBeltframeTemp(reading model.HeatexReading) float64 {
	base := 300.0
	productHeat := reading.ProductFlowTPH * 50
	airCool := reading.AirflowTPH * 2
	return base + productHeat - airCool
}

func (b *BurnerController) PasteurStable(reading model.HeatexReading) bool {
	if reading.BurnerPhase != model.BurnerStable && reading.BurnerPhase != model.BurnerIgnition {
		return false
	}
	return reading.BeltframeTempF > 800 && reading.ExcessO2Pct >= model.MinBeltframeO2Percent
}

func (b *BurnerController) TripRequired(reading model.HeatexReading) bool {
	if reading.ExcessO2Pct > model.MaxBeltframeO2Percent*2 {
		return true
	}
	if reading.BurnerPhase == model.BurnerTrip {
		return true
	}
	if reading.BeltframeTempF > 3500 {
		return true
	}
	return false
}

func (b *BurnerController) PhaseLabel(phase model.BurnerPhase) string {
	switch phase {
	case model.BurnerIdle:
		return "Idle"
	case model.BurnerRinse:
		return "Rinse"
	case model.BurnerIgnition:
		return "Ignition"
	case model.BurnerStable:
		return "Stable Pasteur"
	case model.BurnerTrip:
		return "Tripped"
	default:
		return string(phase)
	}
}

func (b *BurnerController) HeatReleaseMW(reading model.HeatexReading) float64 {
	return reading.ProductFlowTPH * 12.5
}

func (b *BurnerController) TurndownRatio(settings model.PlantSettings, currentProduct float64) float64 {
	if settings.ProductFlowTPH <= 0 {
		return 0
	}
	return currentProduct / settings.ProductFlowTPH
}

func (b *BurnerController) MinStableProduct(settings model.PlantSettings) float64 {
	return settings.ProductFlowTPH * 0.25
}

func (b *BurnerController) NormalizeProduct(flow, max float64) float64 {
	return math.Min(math.Max(flow, 0), max)
}
