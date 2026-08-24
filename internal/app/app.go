package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/lacsar712/pasteur/internal/tunnel"
	"github.com/lacsar712/pasteur/internal/clock"
	"github.com/lacsar712/pasteur/internal/heatex"
	"github.com/lacsar712/pasteur/internal/config"
	"github.com/lacsar712/pasteur/internal/carrier"
	"github.com/lacsar712/pasteur/internal/fsm"
	"github.com/lacsar712/pasteur/internal/interlock"
	"github.com/lacsar712/pasteur/internal/model"
	"github.com/lacsar712/pasteur/internal/store"
)

type App struct {
	cfg           config.Config
	clk           clock.ProcessClock
	store         *store.PlantStore
	journal       *store.Journal
	fsm           *fsm.TunnelFSM
	tunnel        *tunnel.Controller
	heatex    *heatex.Coordinator
	carrier          *carrier.Coordinator
	interlock     *interlock.Interlock
	permissives   *interlock.PermissiveSet
	coordLock     *interlock.CoordinationLock
	scheduler     *clock.Scheduler
	rinseWindow   *clock.RinseWindow
	warmupWindow  *clock.HeatexWarmupWindow
	telemetry     *Telemetry
	tickCancels    map[string]context.CancelFunc
	productLoopCancels map[string]context.CancelFunc
	mu             sync.RWMutex
}

func New(cfg config.Config, clk clock.ProcessClock) *App {
	return &App{
		cfg:          cfg,
		clk:          clk,
		store:        store.NewPlantStore(),
		journal:      store.NewJournal(cfg.JournalPath, cfg.JournalCapacity),
		fsm:          fsm.NewTunnelFSM(cfg.UnitID),
		tunnel:       tunnel.NewController(clk),
		heatex:   heatex.NewCoordinator(clk),
		carrier:         carrier.NewCoordinator(clk),
		interlock:    interlock.NewInterlock(cfg.LeaseTTL),
		permissives:  interlock.NewPermissiveSet(),
		coordLock:    interlock.NewCoordinationLock(),
		scheduler:    clock.NewScheduler(clk),
		rinseWindow:  clock.NewRinseWindow(clk),
		warmupWindow: clock.NewHeatexWarmupWindow(clk),
		telemetry:    NewTelemetry(cfg.UnitID),
		tickCancels:     make(map[string]context.CancelFunc),
		productLoopCancels: make(map[string]context.CancelFunc),
	}
}

func (a *App) Snapshot() model.PlantSnapshot {
	snap, err := a.store.Require(a.cfg.UnitID)
	if err != nil {
		return model.DefaultSnapshot(a.cfg.UnitID)
	}
	return snap
}

func (a *App) Config() config.Config              { return a.cfg }
func (a *App) Clock() clock.ProcessClock          { return a.clk }
func (a *App) FSM() *fsm.TunnelFSM                { return a.fsm }
func (a *App) UnitID() string                     { return a.cfg.UnitID }
func (a *App) Store() *store.PlantStore           { return a.store }
func (a *App) Interlock() *interlock.Interlock    { return a.interlock }
func (a *App) Telemetry() TelemetrySnapshot       { return a.telemetry.Snapshot() }
func (a *App) Journal() *store.Journal            { return a.journal }

func (a *App) journalEvent(ev, payload string) {
	_, _ = a.journal.Append(a.cfg.UnitID, ev, payload)
}

func (a *App) syncState(state model.PlantState) {
	_ = a.store.UpdateState(a.cfg.UnitID, state)
}

func (a *App) isFiring(state model.PlantState) bool {
	return state == model.StateFiring || state == model.StateLoadFollow || state == model.StateRamp
}

func (a *App) refreshPermissives(snap model.PlantSnapshot) {
	a.permissives.SetCarrier(a.carrier.Level().WithinLimits(snap.Carrier.LevelPercent))
	a.permissives.SetPressure(a.tunnel.Pressure().WithinTripLimits(snap.Tunnel.SteamPressurePSI, a.isFiring(snap.State)))
	a.permissives.SetHeatex(a.heatex.Burner().PasteurStable(snap.Heatex))
	a.permissives.SetProduct(snap.Heatex.ProductFlowTPH > 0 || snap.State == model.StateRinse)
	a.permissives.SetIgnition(snap.Heatex.BurnerPhase == model.BurnerStable || snap.Heatex.BurnerPhase == model.BurnerIgnition)
	a.fsm.SetProductPermissive(a.permissives.ProductOK())
	a.fsm.SetRinseComplete(a.rinseWindow.Ready(snap.Heatex.RinseStartedAt))
}

func (a *App) tickLabel() string {
	return fmt.Sprintf("%s-tick", a.cfg.UnitID)
}
