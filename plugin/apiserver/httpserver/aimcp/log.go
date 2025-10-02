package aimcp

import (
	commonlog "github.com/pole-io/pole-server/pkg/common/log"
)

var (
	log = commonlog.GetScopeOrDefaultByName(commonlog.APIServerLoggerName)
)
