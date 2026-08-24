package carrier

import (
	"math"

	"github.com/lacsar712/pasteur/internal/clock"
	"github.com/lacsar712/pasteur/internal/model"
)

type LevelController struct {
	clk clock.ProcessClock
}

func NewLevelController(clk clock.ProcessClock) *LevelController {
	return &LevelController{clk: clk}
}

func (l *LevelController) Compute(snap model.PlantSnapshot, firing bool) (float64, model.CarrierCondition) {
	level := snap.Carrier.LevelPercent
	if !firing {
		return level, model.CarrierNormal
	}
	balance := snap.Carrier.FeedwaterTPH - snap.Carrier.SteamFlowTPH
	level += balance * 0.01
	level = math.Max(model.MinCarrierLevelPercent, math.Min(model.MaxCarrierLevelPercent, level))
	cond := l.classify(level, snap)
	return level, cond
}

func (l *LevelController) classify(level float64, snap model.PlantSnapshot) model.CarrierCondition {
	setpoint := snap.Settings.CarrierLevelSetpoint
	if level > setpoint+15 {
		return model.CarrierSwell
	}
	if level < setpoint-15 {
		return model.CarrierShrink
	}
	if snap.Tunnel.SteamPressurePSI > snap.Settings.TargetSteamPSI*0.9 && level > setpoint+5 {
		return model.CarrierCarry
	}
	return model.CarrierNormal
}

func (l *LevelController) RecommendFeedwater(snap model.PlantSnapshot, firing bool) float64 {
	if !firing {
		return 0
	}
	err := snap.Settings.CarrierLevelSetpoint - snap.Carrier.LevelPercent
	return snap.Settings.FeedwaterFlowTPH + err*3
}

func (l *LevelController) WithinLimits(level float64) bool {
	return level >= model.MinCarrierLevelPercent && level <= model.MaxCarrierLevelPercent
}

func (l *LevelController) TripLow(level float64) bool  { return level < model.TripCarrierLowPercent }
func (l *LevelController) TripHigh(level float64) bool { return level > model.TripCarrierHighPercent }

func (l *LevelController) LevelError(snap model.PlantSnapshot) float64 {
	return snap.Carrier.LevelPercent - snap.Settings.CarrierLevelSetpoint
}

func (l *LevelController) ThreeElementBias(snap model.PlantSnapshot) float64 {
	steam := snap.Carrier.SteamFlowTPH
	feed := snap.Carrier.FeedwaterTPH
	levelErr := l.LevelError(snap)
	return feed + (steam-feed)*0.5 + levelErr*2
}
