package poleagent

import (
	"errors"
	"fmt"
)

type ErrorCategory string

const (
	CategoryInvalidRequest     ErrorCategory = "INVALID_REQUEST"
	CategoryRuntimeUnavailable ErrorCategory = "RUNTIME_UNAVAILABLE"
	CategoryModelUnavailable   ErrorCategory = "MODEL_UNAVAILABLE"
	CategoryToolUnavailable    ErrorCategory = "TOOL_UNAVAILABLE"
	CategoryToolRejected       ErrorCategory = "TOOL_REJECTED"
	CategoryToolLoopLimit      ErrorCategory = "TOOL_LOOP_LIMIT"
	CategoryDownstreamFailed   ErrorCategory = "DOWNSTREAM_FAILED"
	CategoryValidationFailed   ErrorCategory = "VALIDATION_FAILED"
	CategoryNotDraftable       ErrorCategory = "NOT_DRAFTABLE"
)

var (
	ErrInvalidRequest     = errors.New("invalid agent request")
	ErrRuntimeUnavailable = errors.New("agent runtime is unavailable")
	ErrToolLoopLimit      = errors.New("agent tool loop limit reached")
)

type RuntimeError struct {
	Category  ErrorCategory `json:"category"`
	Code      uint32        `json:"code"`
	Info      string        `json:"info"`
	RequestID string        `json:"requestId,omitempty"`
	Retryable bool          `json:"retryable"`
	Cause     error         `json:"-"`
}

func (e *RuntimeError) Error() string {
	if e == nil {
		return ""
	}
	if e.RequestID == "" {
		return fmt.Sprintf("%s: %s", e.Category, e.Info)
	}
	return fmt.Sprintf("%s: %s (requestId=%s)", e.Category, e.Info, e.RequestID)
}

func (e *RuntimeError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func runtimeError(category ErrorCategory, code uint32, info string, retryable bool, cause error) *RuntimeError {
	return &RuntimeError{Category: category, Code: code, Info: info, Retryable: retryable, Cause: cause}
}
