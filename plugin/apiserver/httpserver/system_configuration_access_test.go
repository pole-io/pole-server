package httpserver

import (
	"context"
	"errors"
	"testing"

	"github.com/pole-io/pole-server/apis/pkg/types"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
)

type systemConfigurationCredentialChecker struct {
	err       error
	pass      bool
	module    authtypes.BzModule
	operation authtypes.ResourceOperation
	methods   []authtypes.ServerFunctionName
}

func (c *systemConfigurationCredentialChecker) CheckConsolePermission(authCtx *authtypes.AcquireContext) (bool, error) {
	c.module = authCtx.GetModule()
	c.operation = authCtx.GetOperation()
	c.methods = authCtx.GetMethods()
	return c.pass, c.err
}

func (c *systemConfigurationCredentialChecker) CheckCredential(authCtx *authtypes.AcquireContext) error {
	c.module = authCtx.GetModule()
	c.operation = authCtx.GetOperation()
	c.methods = authCtx.GetMethods()
	return c.err
}

func TestAuthorizeSystemConfigurationRequiresMaintainReadCredential(t *testing.T) {
	checker := &systemConfigurationCredentialChecker{pass: true}
	ctx := context.WithValue(context.Background(), types.ContextAuthTokenKey, "valid-token")
	if err := authorizeSystemConfiguration(checker, checker, ctx); err != nil {
		t.Fatalf("authorizeSystemConfiguration() error = %v", err)
	}
	if checker.module != authtypes.MaintainModule || checker.operation != authtypes.Read {
		t.Fatalf("authorization scope = (%v, %v)", checker.module, checker.operation)
	}
	if len(checker.methods) != 1 || checker.methods[0] != authtypes.DescribeSystemConfiguration {
		t.Fatalf("authorization methods = %v", checker.methods)
	}
}

func TestAuthorizeSystemConfigurationRejectsInvalidCredential(t *testing.T) {
	want := errors.New("invalid credential")
	checker := &systemConfigurationCredentialChecker{err: want}
	if err := authorizeSystemConfiguration(checker, checker, context.Background()); !errors.Is(err, want) {
		t.Fatalf("authorizeSystemConfiguration() error = %v, want %v", err, want)
	}
}

func TestAuthorizeSystemConfigurationRejectsPolicyDenial(t *testing.T) {
	checker := &systemConfigurationCredentialChecker{}
	if err := authorizeSystemConfiguration(checker, checker, context.Background()); !errors.Is(err, errSystemConfigurationPermissionDenied) {
		t.Fatalf("authorizeSystemConfiguration() error = %v", err)
	}
}

func TestAuthorizeSystemConfigurationRejectsMissingChecker(t *testing.T) {
	if err := authorizeSystemConfiguration(nil, nil, context.Background()); err == nil {
		t.Fatal("authorizeSystemConfiguration() error = nil")
	}
}
