package batchctrl

import (
	"context"
	"sync/atomic"
	"time"
)

type Param any

type Future interface {
	Param() Param
	Done() (any, error)
	DoneTimeout(timeout time.Duration) (any, error)
	Cancel()
	Reply(any, error)
	// Attach and Detach are used to attach and detach additional data to the future.
	Attach(string, any)
	Detach(string) any
}

type errorFuture struct {
	task Param
	err  error
}

func (f *errorFuture) Param() Param {
	return f.task
}

func (f *errorFuture) Done() (interface{}, error) {
	return nil, f.err
}

func (f *errorFuture) DoneTimeout(timeout time.Duration) (interface{}, error) {
	return nil, f.err
}

func (f *errorFuture) Cancel() {
}

func (f *errorFuture) Reply(result interface{}, err error) {

}

func (f *errorFuture) Attach(key string, value any) {
	// No-op for errorFuture
}

func (f *errorFuture) Detach(key string) any {
	// No-op for errorFuture
	return nil
}

type future struct {
	task        Param
	setsignal   chan struct{}
	err         error
	result      any
	replied     int32
	ctx         context.Context
	cancel      context.CancelFunc
	attachments map[string]any
}

func (f *future) Param() Param {
	return f.task
}

func (f *future) Done() (interface{}, error) {
	select {
	case <-f.ctx.Done():
		return nil, f.ctx.Err()
	case <-f.setsignal:
		f.cancel()
		return f.result, f.err
	}
}

func (f *future) DoneTimeout(timeout time.Duration) (interface{}, error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil, context.DeadlineExceeded
	case <-f.ctx.Done():
		return nil, f.ctx.Err()
	case <-f.setsignal:
		f.cancel()
		return f.result, f.err
	}
}

func (f *future) Cancel() {
	if atomic.CompareAndSwapInt32(&f.replied, 0, 1) {
		close(f.setsignal)
	}
	f.cancel()
}

func (f *future) Reply(result interface{}, err error) {
	if !atomic.CompareAndSwapInt32(&f.replied, 0, 1) {
		return
	}
	f.result = result
	f.err = err
	close(f.setsignal)
}

func (f *future) Attach(key string, value any) {
	if f.attachments == nil {
		f.attachments = make(map[string]any)
	}
	f.attachments[key] = value
}

func (f *future) Detach(key string) any {
	if f.attachments == nil {
		return nil
	}
	value, exists := f.attachments[key]
	if exists {
		delete(f.attachments, key)
		return value
	}
	return nil
}
