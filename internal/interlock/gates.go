package interlock

import (
	"fmt"

	"github.com/lacsar712/pasteur/internal/model"
)

type PermissiveSet struct {
	productOK       bool
	ignitionOK   bool
	carrierOK       bool
	pressureOK   bool
	heatexOK bool
}

func NewPermissiveSet() *PermissiveSet { return &PermissiveSet{} }

func (p *PermissiveSet) SetProduct(ok bool)       { p.productOK = ok }
func (p *PermissiveSet) SetIgnition(ok bool)   { p.ignitionOK = ok }
func (p *PermissiveSet) SetCarrier(ok bool)       { p.carrierOK = ok }
func (p *PermissiveSet) SetPressure(ok bool)   { p.pressureOK = ok }
func (p *PermissiveSet) SetHeatex(ok bool) { p.heatexOK = ok }

func (p *PermissiveSet) ProductOK() bool       { return p.productOK }
func (p *PermissiveSet) IgnitionOK() bool   { return p.ignitionOK }
func (p *PermissiveSet) CarrierOK() bool       { return p.carrierOK }
func (p *PermissiveSet) PressureOK() bool   { return p.pressureOK }
func (p *PermissiveSet) HeatexOK() bool { return p.heatexOK }

func (p *PermissiveSet) AllFiring() bool {
	return p.productOK && p.ignitionOK && p.carrierOK && p.pressureOK && p.heatexOK
}

func (p *PermissiveSet) CheckIgnition() error {
	if !p.productOK {
		return fmt.Errorf("%w", model.ErrProductPermissive)
	}
	if !p.ignitionOK {
		return fmt.Errorf("%w", model.ErrIgnitionBlocked)
	}
	return nil
}

func CheckPasteurLoss(reading model.HeatexReading) error {
	if reading.BurnerPhase == model.BurnerStable && reading.BeltframeTempF < 600 {
		return fmt.Errorf("%w", model.ErrPasteurLoss)
	}
	return nil
}

func (p *PermissiveSet) CheckFiring() error {
	if err := p.CheckIgnition(); err != nil {
		return err
	}
	if !p.carrierOK {
		return fmt.Errorf("%w", model.ErrCarrierLevelTrip)
	}
	if !p.pressureOK {
		return fmt.Errorf("%w", model.ErrPressureTrip)
	}
	if !p.heatexOK {
		return fmt.Errorf("%w", model.ErrHeatexTrip)
	}
	return nil
}

type CoordinationLock struct {
	holder string
	held   bool
}

func NewCoordinationLock() *CoordinationLock { return &CoordinationLock{} }

func (c *CoordinationLock) Acquire(holder string) error {
	if c.held {
		return fmt.Errorf("%w", model.ErrCoordinationLock)
	}
	c.holder = holder
	c.held = true
	return nil
}

func (c *CoordinationLock) Release(holder string) {
	if c.held && c.holder == holder {
		c.held = false
		c.holder = ""
	}
}

func (c *CoordinationLock) Require(holder string) error {
	if !c.held || c.holder != holder {
		return fmt.Errorf("%w", model.ErrCoordinationLock)
	}
	return nil
}

func (c *CoordinationLock) Held() bool { return c.held }
