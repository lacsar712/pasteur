package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lacsar712/pasteur/internal/app"
	"github.com/lacsar712/pasteur/internal/clock"
	"github.com/lacsar712/pasteur/internal/config"
	"github.com/lacsar712/pasteur/internal/model"
)

func TestCase(t *testing.T) {
	clk := clock.NewManual(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	cfg := config.Default("BD-1")
	a, err := app.BootstrapWithClock(cfg, clk)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Trip(context.Background(), "test"); err != nil {
		t.Fatal(err)
	}
	err = a.DrainAfterShutdown(context.Background(), 100)
	if err == nil {
		t.Fatal("expected blowdown limit error")
	}
	if !errors.Is(err, model.ErrDrainLimit) {
		t.Fatalf("expected ErrDrainLimit, got %v", err)
	}
}
