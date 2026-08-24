package fsm

import (
	"context"
	"fmt"
	"sync"

	"github.com/lacsar712/pasteur/internal/model"
)

type TunnelFSM struct {
	mu            sync.RWMutex
	state         model.PlantState
	productPermissive bool
	rinseComplete  bool
	hooks          *HookChain
}

func NewTunnelFSM(unitID string) *TunnelFSM {
	_ = unitID
	return &TunnelFSM{state: model.StateColdStandby, hooks: NewHookChain()}
}

func (f *TunnelFSM) Hooks() *HookChain { return f.hooks }

func (f *TunnelFSM) State() model.PlantState {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.state
}

func (f *TunnelFSM) SetProductPermissive(ok bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.productPermissive = ok
}

func (f *TunnelFSM) SetRinseComplete(ok bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rinseComplete = ok
}

func (f *TunnelFSM) ProductPermissive() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.productPermissive
}

func (f *TunnelFSM) Dispatch(ctx context.Context, event PlantEvent) (model.PlantState, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	select {
	case <-ctx.Done():
		return f.state, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	if event == EvTrip {
		from := f.state
		if f.hooks != nil {
			if err := f.hooks.RunBefore(ctx, from, model.StateTrip, event); err != nil {
				return f.state, err
			}
		}
		f.state = model.StateTrip
		if f.hooks != nil {
			if err := f.hooks.RunAfter(ctx, from, model.StateTrip, event); err != nil {
				return f.state, err
			}
		}
		return f.state, nil
	}
	next, ok := NextState(f.state, event)
	if !ok {
		// Rejected transition: no state change, so acceptance side-effects
		// (e.g. the beltframe transport pulse) must not fire. Returning here
		// without RunAfter keeps the bypass exit from driving the execution
		// chain while the belt is parked in standby.
		return f.state, fmt.Errorf("%s from %s: %w", event, f.state, ErrIllegalTransition)
	}
	if event == EvIgnite && !f.productPermissive {
		return f.state, fmt.Errorf("%w", model.ErrProductPermissive)
	}
	if event == EvRinseComplete && !f.rinseComplete {
		return f.state, fmt.Errorf("%w", model.ErrRinseIncomplete)
	}
	from := f.state
	if f.hooks != nil {
		if err := f.hooks.RunBefore(ctx, from, next, event); err != nil {
			return f.state, err
		}
	}
	f.state = next
	if f.hooks != nil {
		if err := f.hooks.RunAfter(ctx, from, next, event); err != nil {
			return f.state, err
		}
	}
	return f.state, nil
}

func (f *TunnelFSM) ForceState(state model.PlantState) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state = state
}
