package supervisor

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeModule struct {
	name     string
	events   *[]string
	mu       *sync.Mutex
	startErr error
	runErr   chan error
	stopErr  error
}

func (f *fakeModule) Name() string {
	return f.name
}

func (f *fakeModule) Start(context.Context) (Running, error) {
	f.record("start:" + f.name)
	if f.startErr != nil {
		return nil, f.startErr
	}
	return &fakeRunning{module: f}, nil
}

func (f *fakeModule) record(event string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	*f.events = append(*f.events, event)
}

type fakeRunning struct {
	module *fakeModule
}

func (f *fakeRunning) Wait() error {
	return <-f.module.runErr
}

func (f *fakeRunning) Stop(context.Context) error {
	f.module.record("stop:" + f.module.name)
	return f.module.stopErr
}

func TestSupervisorStartsAndStopsInDependencyOrder(t *testing.T) {
	events := []string{}
	var mu sync.Mutex
	modules := []*fakeModule{
		{name: "control-plane", events: &events, mu: &mu, runErr: make(chan error, 1)},
		{name: "limiter-server", events: &events, mu: &mu, runErr: make(chan error, 1)},
		{name: "console", events: &events, mu: &mu, runErr: make(chan error, 1)},
	}
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- Supervisor{
			Modules: []Module{modules[0], modules[1], modules[2]},
		}.Run(ctx)
	}()
	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(events) == 3
	}, time.Second, time.Millisecond)
	cancel()
	err := <-errCh

	require.NoError(t, err)
	require.Equal(t, []string{
		"start:control-plane",
		"start:limiter-server",
		"start:console",
		"stop:console",
		"stop:limiter-server",
		"stop:control-plane",
	}, events)
}

func TestSupervisorRollsBackEarlierModulesOnStartFailure(t *testing.T) {
	events := []string{}
	var mu sync.Mutex
	controlPlane := &fakeModule{
		name: "control-plane", events: &events, mu: &mu, runErr: make(chan error, 1)}
	limiter := &fakeModule{
		name: "limiter-server", events: &events, mu: &mu,
		startErr: errors.New("bind failed"), runErr: make(chan error, 1)}

	err := Supervisor{Modules: []Module{controlPlane, limiter}}.Run(context.Background())

	require.ErrorContains(t, err, "module limiter-server start: bind failed")
	require.Equal(t, []string{
		"start:control-plane",
		"start:limiter-server",
		"stop:control-plane",
	}, events)
}

func TestSupervisorFailsFastOnRuntimeError(t *testing.T) {
	events := []string{}
	var mu sync.Mutex
	controlPlane := &fakeModule{
		name: "control-plane", events: &events, mu: &mu, runErr: make(chan error, 1)}
	limiter := &fakeModule{
		name: "limiter-server", events: &events, mu: &mu, runErr: make(chan error, 1)}
	limiter.runErr <- errors.New("serve failed")

	err := Supervisor{
		Modules:         []Module{controlPlane, limiter},
		ShutdownTimeout: time.Second,
	}.Run(context.Background())

	require.ErrorContains(t, err, "module limiter-server run: serve failed")
	require.Equal(t, "stop:control-plane", events[len(events)-1])
}
