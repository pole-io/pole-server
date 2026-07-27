package supervisor

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Module 是可由进程编排器管理的运行模块。
type Module interface {
	Name() string
	Start(context.Context) (Running, error)
}

// Running 是已经完成 readiness 的运行实例。
type Running interface {
	Wait() error
	Stop(context.Context) error
}

// Supervisor 串行启动模块，并在任一模块异常时逆序停止全部实例。
type Supervisor struct {
	Modules         []Module
	ShutdownTimeout time.Duration
}

// Run 运行选定 Profile。父 Context 正常取消返回 nil，模块错误返回带阶段上下文的错误。
func (s Supervisor) Run(parent context.Context) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	timeout := s.ShutdownTimeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	type startedModule struct {
		name    string
		running Running
	}
	started := make([]startedModule, 0, len(s.Modules))
	runErrors := make(chan error, 1)
	stopping := make(chan struct{})

	watch := func(name string, running Running) {
		go func() {
			err := running.Wait()
			select {
			case <-stopping:
				return
			default:
			}
			if err == nil {
				err = errors.New("exited unexpectedly")
			}
			wrapped := fmt.Errorf("module %s run: %w", name, err)
			select {
			case runErrors <- wrapped:
				cancel()
			default:
			}
		}()
	}

	var mainErr error
startLoop:
	for _, module := range s.Modules {
		select {
		case <-ctx.Done():
			break startLoop
		default:
		}
		running, err := module.Start(ctx)
		if err != nil {
			mainErr = fmt.Errorf("module %s start: %w", module.Name(), err)
			break
		}
		started = append(started, startedModule{name: module.Name(), running: running})
		watch(module.Name(), running)
		select {
		case err := <-runErrors:
			mainErr = err
		default:
		}
		if mainErr != nil {
			break
		}
	}

	if mainErr == nil {
		select {
		case mainErr = <-runErrors:
		case <-parent.Done():
		}
	}

	close(stopping)
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), timeout)
	defer shutdownCancel()
	var stopErrors []error
	for i := len(started) - 1; i >= 0; i-- {
		if err := started[i].running.Stop(shutdownCtx); err != nil {
			stopErrors = append(stopErrors,
				fmt.Errorf("module %s stop: %w", started[i].name, err))
		}
	}
	if len(stopErrors) == 0 {
		return mainErr
	}
	return errors.Join(append([]error{mainErr}, stopErrors...)...)
}
