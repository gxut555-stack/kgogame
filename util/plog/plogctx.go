package plog

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"kgogame/util/gcontext"
)

// 带调用函数名和context的Info
func CInfof(ctx context.Context, format string, t ...interface{}) {
	Init()
	file, line, fn := getCaller(2)
	sl := []zap.Field{
		zap.String("FI", file),
		zap.Int("LN", line),
	}
	traceId, ok := ctx.Value(gcontext.TraceId).(string)
	if ok {
		fn = traceId + "_" + fn
	}
	Lzap.With(sl...).Info(fmt.Sprintf(fn+"|"+format, t...))
}

// 带调用函数名和context的Debug
func CDebugf(ctx context.Context, format string, t ...interface{}) {
	Init()
	file, line, fn := getCaller(2)
	sl := []zap.Field{
		zap.String("FI", file),
		zap.Int("LN", line),
	}
	traceId, ok := ctx.Value(gcontext.TraceId).(string)
	if ok {
		fn = traceId + "_" + fn
	}
	Lzap.With(sl...).Debug(fmt.Sprintf(fn+"|"+format, t...))
}

// 带调用函数名和context的Warn
func CWarnf(ctx context.Context, format string, t ...interface{}) {
	Init()
	file, line, fn := getCaller(2)
	sl := []zap.Field{
		zap.String("FI", file),
		zap.Int("LN", line),
	}
	traceId, ok := ctx.Value(gcontext.TraceId).(string)
	if ok {
		fn = traceId + "_" + fn
	}
	Lzap.With(sl...).Warn(fmt.Sprintf(fn+"|"+format, t...))
}

// 带调用函数名和context的Fatal
func CFatalf(ctx context.Context, format string, t ...interface{}) {
	Init()
	file, line, fn := getCaller(2)
	sl := []zap.Field{
		zap.String("FI", file),
		zap.Int("LN", line),
	}
	traceId, ok := ctx.Value(gcontext.TraceId).(string)
	if ok {
		fn = traceId + "_" + fn
	}
	Lzap.With(sl...).Error(fmt.Sprintf(fn+"|"+format, t...))
}
