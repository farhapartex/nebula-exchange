package health

import "context"

type DependencyProbe interface {
	Name() string
	Ping(ctx context.Context) error
}

type probeFunction struct {
	name string
	ping func(ctx context.Context) error
}

func NewProbe(name string, ping func(ctx context.Context) error) DependencyProbe {
	return probeFunction{name: name, ping: ping}
}

func (probe probeFunction) Name() string {
	return probe.name
}

func (probe probeFunction) Ping(ctx context.Context) error {
	return probe.ping(ctx)
}
