package model

import "errors"

var (
	ErrContextDone      = errors.New("operation cancelled")
	ErrPlantNotFound    = errors.New("plant unit not found")
	ErrLeaseHeld        = errors.New("interlock lease held by another operator")
	ErrLeaseMissing     = errors.New("interlock lease missing or expired")
	ErrGateBlocked      = errors.New("safety gate blocked")
	ErrProductPermissive   = errors.New("product permissive not satisfied")
	ErrIgnitionBlocked  = errors.New("ignition sequence blocked")
	ErrCarrierLevelTrip    = errors.New("carrier level trip condition")
	ErrPressureTrip     = errors.New("steam pressure trip condition")
	ErrHeatexTrip   = errors.New("heatex trip condition")
	ErrIllegalState     = errors.New("illegal plant state transition")
	ErrSnapshotStale    = errors.New("snapshot revision stale")
	ErrWindowOpen       = errors.New("timing window still open")
	ErrRinseIncomplete  = errors.New("beltframe rinse incomplete")
	ErrCoordinationLock = errors.New("coordination lock held")
	ErrCarrierLevelLow     = errors.New("carrier level below low limit")
	ErrPasteurLoss        = errors.New("beltframe pasteur lost")
	ErrDrainLimit    = errors.New("drain valve at limit")
)
