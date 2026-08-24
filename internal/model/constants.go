package model

import "time"

const (
	DefaultLeaseTTL        = 30 * time.Second
	RinseWindow            = 5 * time.Minute
	IgnitionDelayWindow    = 15 * time.Second
	CarrierSwellSettleWindow  = 45 * time.Second
	HeatexWarmupWindow = 2 * time.Minute
	FeedwaterRampWindow    = 30 * time.Second
	MaxCarrierLevelPercent    = 95.0
	MinCarrierLevelPercent    = 15.0
	TripCarrierLowPercent     = 10.0
	TripCarrierHighPercent    = 98.0
	NormalSteamPressurePSI = 1800.0
	MaxSteamPressurePSI    = 2000.0
	MinBeltframeO2Percent    = 2.5
	MaxBeltframeO2Percent    = 6.0
	DefaultJournalCapacity = 512
)
