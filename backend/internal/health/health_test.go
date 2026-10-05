package health

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"
)

func TestCheckReportsEveryDependency(t *testing.T) {
	quietLogger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	healthyProbe := NewProbe("database", func(context.Context) error { return nil })
	failingProbe := NewProbe("redis", func(context.Context) error { return errors.New("connection refused") })

	healthyReport := NewService(quietLogger, healthyProbe).Check(context.Background())
	degradedReport := NewService(quietLogger, healthyProbe, failingProbe).Check(context.Background())

	if !healthyReport.IsHealthy() || healthyReport.Dependencies["database"] != StatusOK {
		t.Fatalf("expected a healthy report, got %+v", healthyReport)
	}
	if degradedReport.IsHealthy() || degradedReport.Dependencies["redis"] != StatusUnavailable || degradedReport.Dependencies["database"] != StatusOK {
		t.Fatalf("expected redis to be reported unavailable, got %+v", degradedReport)
	}
}
