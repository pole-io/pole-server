package apolloserver

import (
	commonlog "github.com/pole-io/pole-server/pkg/common/log"
)

var (
	apollolog = commonlog.RegisterScope("apollo-apiserver", "apollo apiserver plugin", 0)
	tracelog  = commonlog.RegisterScope("apollo-trace", "apollo trace", 0)
)
