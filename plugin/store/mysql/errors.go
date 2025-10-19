package sqldb

import "errors"

var (
	ErrorMissingParams = errors.New("missing some params")
	ErrTxIsNil         = errors.New("tx is nil")
)
