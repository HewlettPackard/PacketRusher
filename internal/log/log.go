/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */

// Package log is PacketRusher's logger: zap behind the few functions the code uses.
package log

import (
	"os"
	"sync/atomic"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	level  = zap.NewAtomicLevelAt(zap.InfoLevel)
	logger atomic.Pointer[zap.SugaredLogger]
)

func init() {
	encoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey: "time", LevelKey: "level", MessageKey: "msg",
		EncodeTime: zapcore.ISO8601TimeEncoder, EncodeLevel: zapcore.LowercaseLevelEncoder,
	})
	logger.Store(zap.New(zapcore.NewCore(encoder, zapcore.Lock(os.Stdout), level)).Sugar())
}

// SetLevel takes the level of the configuration file: 2 or less logs errors
// only, 3 adds warnings, 4 (the default) information and 5 or more debugging.
func SetLevel(configured int) {
	switch {
	case configured >= 5:
		level.SetLevel(zap.DebugLevel)
	case configured == 4 || configured == 0:
		level.SetLevel(zap.InfoLevel)
	case configured == 3:
		level.SetLevel(zap.WarnLevel)
	default:
		level.SetLevel(zap.ErrorLevel)
	}
}

// Replace sends the log entries to core instead of the standard output until the
// returned function is called. Tests use it to observe what is logged.
func Replace(core zapcore.Core) (restore func()) {
	previous := logger.Swap(zap.New(core).Sugar())
	return func() { logger.Store(previous) }
}

func Debug(args ...any)                 { logger.Load().Debug(args...) }
func Debugf(format string, args ...any) { logger.Load().Debugf(format, args...) }
func Info(args ...any)                  { logger.Load().Info(args...) }
func Infof(format string, args ...any)  { logger.Load().Infof(format, args...) }
func Warn(args ...any)                  { logger.Load().Warn(args...) }
func Warnf(format string, args ...any)  { logger.Load().Warnf(format, args...) }
func Error(args ...any)                 { logger.Load().Error(args...) }
func Errorf(format string, args ...any) { logger.Load().Errorf(format, args...) }
func Fatal(args ...any)                 { logger.Load().Fatal(args...) }
func Fatalf(format string, args ...any) { logger.Load().Fatalf(format, args...) }
