package utils

import (
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

// NewUUID 返回一个随机的UUID
func NewUUID() string {
	uuidBytes := uuid.New()
	return hex.EncodeToString(uuidBytes[:])
}

// Int2bool 整数转换为bool值
func Int2bool(entry int) bool {
	return entry != 0
}

// StatusBoolToInt 状态bool转int
func StatusBoolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

// Time2String Convert time.Time to string time
func Time2String(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// Int64Time2String Convert time stamp of Int64 to string time
func Int64Time2String(t int64) string {
	return time.Unix(t, 0).Format("2006-01-02 15:04:05")
}

func DefaultString(v, d string) string {
	if v == "" {
		return d
	}
	return v
}
