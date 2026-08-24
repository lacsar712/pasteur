package model

import "time"

func CloneSnapshot(s PlantSnapshot) PlantSnapshot {
	out := s
	out.Alarms = append([]AlarmEvent(nil), s.Alarms...)
	return out
}

func DefaultSnapshot(unitID string) PlantSnapshot {
	now := time.Now()
	return PlantSnapshot{
		UnitID: unitID,
		State:  StateColdStandby,
		Settings: PlantSettings{
			Mode:              ModeBaseLoad,
			TargetMW:          150,
			TargetSteamPSI:    NormalSteamPressurePSI,
			CarrierLevelSetpoint: 55,
			FeedwaterFlowTPH:  400,
			ProductFlowTPH:       35,
			ExcessO2Setpoint:  3.5,
		},
		Plant: PlantRef{UnitLabel: unitID, PlantCode: "STEAM-PLT"},
		Carrier: CarrierReading{
			LevelPercent: 50,
			Condition:    CarrierNormal,
			FeedwaterTPH: 0,
			SteamFlowTPH: 0,
		},
		Heatex: HeatexReading{
			BurnerPhase: BurnerIdle,
		},
		Tunnel: TunnelReading{
			SteamPressurePSI: 0,
			SteamTempF:       70,
		},
		UpdatedAt: now,
	}
}

func (s PlantSnapshot) IsFiring() bool {
	return s.State == StateFiring || s.State == StateLoadFollow || s.State == StateRamp
}

func (s PlantSnapshot) CarrierWithinLimits() bool {
	return s.Carrier.LevelPercent >= MinCarrierLevelPercent && s.Carrier.LevelPercent <= MaxCarrierLevelPercent
}

func (s PlantSnapshot) PressureWithinLimits() bool {
	if !s.IsFiring() {
		return true
	}
	return s.Tunnel.SteamPressurePSI <= MaxSteamPressurePSI
}
