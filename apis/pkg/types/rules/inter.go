package rules

import "time"

type IRule interface {
	GetId() string
	GetMtime() time.Time
}
