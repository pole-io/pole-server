package service

import (
	"testing"
	"time"

	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"
)

func TestServiceInstancesGetInstancesDoesNotHoldLockDuringConsumer(t *testing.T) {
	serviceInstances := NewServiceInstances(0)
	first := CreateInstanceModel("svc-1", &apiservice.Instance{
		Id:      "ins-1",
		Healthy: true,
	})
	second := CreateInstanceModel("svc-1", &apiservice.Instance{
		Id:      "ins-2",
		Healthy: true,
	})
	serviceInstances.UpsertInstance(first)

	done := make(chan struct{})
	go func() {
		serviceInstances.GetInstances(false, func(*Instance) {
			serviceInstances.UpsertInstance(second)
		})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("GetInstances should not hold ServiceInstances lock while invoking consumer")
	}

	foundSecond := false
	serviceInstances.GetInstances(false, func(ins *Instance) {
		if ins.ID() == "ins-2" {
			foundSecond = true
		}
	})
	if !foundSecond {
		t.Fatal("expected snapshot to be rebuilt after instance update")
	}
}
