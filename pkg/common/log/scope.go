/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://opensource.org/licenses/BSD-3-Clause
 *
 * Unless required by applicable law or agreed to in writing, software distributed
 * under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package log

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Scope let's you log data for an area of code, enabling the user full control over
// the level of logging output produced.
type Scope struct {
	// immutable, set at creation
	name        string
	nameToEmit  string
	description string
	callerSkip  int

	// set by the Configure method and adjustable dynamically
	outputLevel     Level
	stackTraceLevel Level
	logCallers      bool
	pt              *patchTable
}

var (
	scopes = map[string]*Scope{}
	lock   sync.RWMutex
)

// RegisterScope registers a new logging scope. If the same name is used multiple times
// for a single process, the same Scope struct is returned.
//
// Scope names cannot include colons, commas, or periods.
func RegisterScope(name string, description string, callerSkip int) *Scope {
	if strings.ContainsAny(name, ":,.") {
		panic(fmt.Sprintf("scope name %s is invalid, it cannot contain colons, commas, or periods", name))
	}

	lock.Lock()
	defer lock.Unlock()

	s, ok := scopes[name]
	if !ok {
		s = &Scope{
			name:        name,
			description: description,
			callerSkip:  callerSkip,
		}
		s.SetOutputLevel(InfoLevel)
		s.SetStackTraceLevel(NoneLevel)
		s.SetDisableLogCaller(true)
		if name != DefaultLoggerName {
			s.nameToEmit = name
		}

		scopes[name] = s
	}

	return s
}

// FindScope returns a previously registered scope, or nil if the named scope wasn't previously registered
func FindScope(scope string) *Scope {
	lock.RLock()
	defer lock.RUnlock()

	s := scopes[scope]
	return s
}

func GetScopeOrDefaultByName(name string) *Scope {
	lock.RLock()
	defer lock.RUnlock()
	s := scopes[name]
	if s == nil {
		s = scopes[DefaultLoggerName]
	}
	return s
}

func GetScopeByName(name, defaultName string) *Scope {
	lock.RLock()
	defer lock.RUnlock()
	s := scopes[name]
	if s == nil {
		s = scopes[defaultName]
		if s == nil {
			s = scopes[DefaultLoggerName]
		}
	}
	return s
}

// Scopes returns a snapshot of the currently defined set of scopes
func Scopes() map[string]*Scope {
	lock.RLock()
	defer lock.RUnlock()

	s := make(map[string]*Scope, len(scopes))
	for k, v := range scopes {
		s[k] = v
	}

	return s
}

// Fatal outputs a message at fatal level.
func (s *Scope) Fatal(msg string, fields ...zapcore.Field) {
	if s.GetOutputLevel() >= FatalLevel {
		s.emit(zapcore.FatalLevel, s.GetStackTraceLevel() >= FatalLevel, msg, fields)
	}
}

// Fatala uses fmt.Sprint to construct and log a message at fatal level.
func (s *Scope) Fatala(args ...interface{}) {
	if s.GetOutputLevel() >= FatalLevel {
		s.emit(zapcore.FatalLevel, s.GetStackTraceLevel() >= FatalLevel, fmt.Sprint(args...), nil)
	}
}

// Fatalf uses fmt.Sprintf to construct and log a message at fatal level.
func (s *Scope) Fatalf(template string, args ...interface{}) {
	if s.GetOutputLevel() >= FatalLevel {
		msg := template
		if len(args) > 0 {
			msg = fmt.Sprintf(template, args...)
		}
		s.emit(zapcore.FatalLevel, s.GetStackTraceLevel() >= FatalLevel, msg, nil)
	}
}

// FatalEnabled returns whether output of messages using this scope is currently enabled for fatal-level output.
func (s *Scope) FatalEnabled() bool {
	return s.GetOutputLevel() >= FatalLevel
}

// Error outputs a message at error level.
func (s *Scope) Error(msg string, fields ...zapcore.Field) {
	if s.GetOutputLevel() >= ErrorLevel {
		s.emit(zapcore.ErrorLevel, s.GetStackTraceLevel() >= ErrorLevel, msg, fields)
	}
}

// Errora uses fmt.Sprint to construct and log a message at error level.
func (s *Scope) Errora(args ...interface{}) {
	if s.GetOutputLevel() >= ErrorLevel {
		s.emit(zapcore.ErrorLevel, s.GetStackTraceLevel() >= ErrorLevel, fmt.Sprint(args...), nil)
	}
}

// Errorf uses fmt.Sprintf to construct and log a message at error level.
func (s *Scope) Errorf(template string, args ...interface{}) {
	if s.GetOutputLevel() >= ErrorLevel {
		msg := template
		if len(args) > 0 {
			msg = fmt.Sprintf(template, args...)
		}
		s.emit(zapcore.ErrorLevel, s.GetStackTraceLevel() >= ErrorLevel, msg, nil)
	}
}

// ErrorEnabled returns whether output of messages using this scope is currently enabled for error-level output.
func (s *Scope) ErrorEnabled() bool {
	return s.GetOutputLevel() >= ErrorLevel
}

// Warn outputs a message at warn level.
func (s *Scope) Warn(msg string, fields ...zapcore.Field) {
	if s.GetOutputLevel() >= WarnLevel {
		s.emit(zapcore.WarnLevel, s.GetStackTraceLevel() >= ErrorLevel, msg, fields)
	}
}

// Warna uses fmt.Sprint to construct and log a message at warn level.
func (s *Scope) Warna(args ...interface{}) {
	if s.GetOutputLevel() >= WarnLevel {
		s.emit(zapcore.WarnLevel, s.GetStackTraceLevel() >= ErrorLevel, fmt.Sprint(args...), nil)
	}
}

// Warnf uses fmt.Sprintf to construct and log a message at warn level.
func (s *Scope) Warnf(template string, args ...interface{}) {
	if s.GetOutputLevel() >= WarnLevel {
		msg := template
		if len(args) > 0 {
			msg = fmt.Sprintf(template, args...)
		}
		s.emit(zapcore.WarnLevel, s.GetStackTraceLevel() >= ErrorLevel, msg, nil)
	}
}

// WarnEnabled returns whether output of messages using this scope is currently enabled for warn-level output.
func (s *Scope) WarnEnabled() bool {
	return s.GetOutputLevel() >= WarnLevel
}

// Info outputs a message at info level.
func (s *Scope) Info(msg string, fields ...zapcore.Field) {
	if s.GetOutputLevel() >= InfoLevel {
		s.emit(zapcore.InfoLevel, s.GetStackTraceLevel() >= ErrorLevel, msg, fields)
	}
}

// Infoa uses fmt.Sprint to construct and log a message at info level.
func (s *Scope) Infoa(args ...interface{}) {
	if s.GetOutputLevel() >= InfoLevel {
		s.emit(zapcore.InfoLevel, s.GetStackTraceLevel() >= ErrorLevel, fmt.Sprint(args...), nil)
	}
}

// Infof uses fmt.Sprintf to construct and log a message at info level.
func (s *Scope) Infof(template string, args ...interface{}) {
	if s.GetOutputLevel() >= InfoLevel {
		msg := template
		if len(args) > 0 {
			msg = fmt.Sprintf(template, args...)
		}
		s.emit(zapcore.InfoLevel, s.GetStackTraceLevel() >= ErrorLevel, msg, nil)
	}
}

// InfoEnabled returns whether output of messages using this scope is currently enabled for info-level output.
func (s *Scope) InfoEnabled() bool {
	return s.GetOutputLevel() >= InfoLevel
}

// Debug outputs a message at debug level.
func (s *Scope) Debug(msg string, fields ...zapcore.Field) {
	if s.GetOutputLevel() >= DebugLevel {
		s.emit(zapcore.DebugLevel, s.GetStackTraceLevel() >= ErrorLevel, msg, fields)
	}
}

// Debuga uses fmt.Sprint to construct and log a message at debug level.
func (s *Scope) Debuga(args ...interface{}) {
	if s.GetOutputLevel() >= DebugLevel {
		s.emit(zapcore.DebugLevel, s.GetStackTraceLevel() >= ErrorLevel, fmt.Sprint(args...), nil)
	}
}

// Debugf uses fmt.Sprintf to construct and log a message at debug level.
func (s *Scope) Debugf(template string, args ...interface{}) {
	if s.GetOutputLevel() >= DebugLevel {
		msg := template
		if len(args) > 0 {
			msg = fmt.Sprintf(template, args...)
		}
		s.emit(zapcore.DebugLevel, s.GetStackTraceLevel() >= ErrorLevel, msg, nil)
	}
}

// DebugEnabled returns whether output of messages using this scope is currently enabled for debug-level output.
func (s *Scope) DebugEnabled() bool {
	return s.GetOutputLevel() >= DebugLevel
}

// Name returns this scope's name.
func (s *Scope) Name() string {
	return s.name
}

// Description returns this scope's description
func (s *Scope) Description() string {
	return s.description
}

func (s *Scope) getPathTable() *patchTable {
	return s.pt
}

const callerSkipOffset = 2

func (s *Scope) emit(level zapcore.Level, dumpStack bool, msg string, fields []zapcore.Field) {
	e := zapcore.Entry{
		Message:    msg,
		Level:      level,
		Time:       time.Now(),
		LoggerName: s.nameToEmit,
	}

	if !s.GetDisableLogCaller() {
		e.Caller = zapcore.NewEntryCaller(runtime.Caller(s.callerSkip + callerSkipOffset))
	}

	if dumpStack {
		e.Stack = zap.Stack("").String
	}

	pt := s.getPathTable()
	if pt != nil && pt.write != nil {
		if err := pt.write(e, fields); err != nil {
			_, _ = fmt.Fprintf(pt.errorSink, "%v log write error: %v\n", time.Now(), err)
			_ = pt.errorSink.Sync()
		}
	}
}

// SetOutputLevel adjusts the output level associated with the scope.
func (s *Scope) SetOutputLevel(l Level) {
	s.outputLevel = l
}

// GetOutputLevel returns the output level associated with the scope.
func (s *Scope) GetOutputLevel() Level {
	return s.outputLevel
}

// SetStackTraceLevel adjusts the stack tracing level associated with the scope.
func (s *Scope) SetStackTraceLevel(l Level) {
	s.stackTraceLevel = l
}

// GetStackTraceLevel returns the stack tracing level associated with the scope.
func (s *Scope) GetStackTraceLevel() Level {
	return s.stackTraceLevel
}

// SetDisableLogCaller adjusts the output level associated with the scope.
func (s *Scope) SetDisableLogCaller(logCallers bool) {
	s.logCallers = logCallers
}

// GetDisableLogCaller returns the output level associated with the scope.
func (s *Scope) GetDisableLogCaller() bool {
	return s.logCallers
}

// Sync 调用log的Sync方法
func (s *Scope) Sync() error {
	pt := s.getPathTable()
	if pt != nil && pt.sync != nil {
		return pt.sync()
	}
	return nil
}
