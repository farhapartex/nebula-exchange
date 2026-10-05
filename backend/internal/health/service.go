package health

import (
	"context"
	"log/slog"
	"time"
)

const (
	StatusOK          = "ok"
	StatusUnavailable = "unavailable"
	probeTimeout      = 2 * time.Second
)

type Report struct {
	Status       string            `json:"status"`
	Dependencies map[string]string `json:"dependencies"`
}

func (report Report) IsHealthy() bool {
	return report.Status == StatusOK
}

type Service interface {
	Check(ctx context.Context) Report
}

type service struct {
	probes []DependencyProbe
	logger *slog.Logger
}

func NewService(logger *slog.Logger, probes ...DependencyProbe) Service {
	return &service{probes: probes, logger: logger}
}

func (healthService *service) Check(ctx context.Context) Report {
	report := Report{Status: StatusOK, Dependencies: make(map[string]string, len(healthService.probes))}
	for _, probe := range healthService.probes {
		probeContext, cancelProbe := context.WithTimeout(ctx, probeTimeout)
		err := probe.Ping(probeContext)
		cancelProbe()
		if err != nil {
			healthService.logger.WarnContext(ctx, "dependency unhealthy", slog.String("dependency", probe.Name()), slog.Any("error", err))
			report.Status = StatusUnavailable
			report.Dependencies[probe.Name()] = StatusUnavailable
			continue
		}
		report.Dependencies[probe.Name()] = StatusOK
	}
	return report
}
