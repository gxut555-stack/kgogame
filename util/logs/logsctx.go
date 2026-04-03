package logs

import (
	"context"
	"kgogame/util/gcontext"
)

// 输出日志到文件中，DEBUG级别, format为格式化配置
func CDebugf(ctx context.Context, format string, v ...interface{}) {
	if !isInitLogOnce {
		InitLog("Uninitialized", 1, nil)
	}
	l := GetInstance()
	if LOG_DEBUG < l.level {
		return
	}

	l.prepare(LOG_DEBUG, false, genCtxFmt(ctx, format), v...)
}

func CInfof(ctx context.Context, format string, v ...interface{}) {
	if !isInitLogOnce {
		InitLog("Uninitialized", 1, nil)
	}
	l := GetInstance()
	if LOG_INFO < l.level {
		return
	}
	l.prepare(LOG_INFO, false, genCtxFmt(ctx, format), v...)
}

func CWarnf(ctx context.Context, format string, v ...interface{}) {
	if !isInitLogOnce {
		InitLog("Uninitialized", 1, nil)
	}
	l := GetInstance()
	if LOG_ERROR < l.level {
		return
	}
	l.prepare(LOG_WARN, false, genCtxFmt(ctx, format), v...)
}

func CErrorf(ctx context.Context, format string, v ...interface{}) {
	if !isInitLogOnce {
		InitLog("Uninitialized", 1, nil)
	}
	l := GetInstance()
	if LOG_ERROR < l.level {
		return
	}
	l.prepare(LOG_ERROR, false, genCtxFmt(ctx, format), v...)
}

func CFatalf(ctx context.Context, format string, v ...interface{}) {
	if !isInitLogOnce {
		InitLog("Uninitialized", 1, nil)
	}
	l := GetInstance()
	if LOG_FATAL < l.level {
		return
	}
	l.prepare(LOG_FATAL, false, genCtxFmt(ctx, format), v...)
}

func genCtxFmt(ctx context.Context, format string) string {
	if ctx == nil {
		return format
	}
	traceId, ok := ctx.Value(gcontext.TraceId).(string)
	if ok {
		return "[" + traceId + "] " + format
	}
	return format
}

// 输出日志到文件中，DEBUG级别, format为格式化配置
func SDebug(format string, v ...interface{}) {
	if !isInitLogOnce {
		InitLog("Uninitialized", 1, nil)
	}
	l := GetInstance()
	if LOG_DEBUG < l.level {
		return
	}

	l.prepare(LOG_DEBUG, false, format, v...)
}

func SInfo(format string, v ...interface{}) {
	if !isInitLogOnce {
		InitLog("Uninitialized", 1, nil)
	}
	l := GetInstance()
	if LOG_INFO < l.level {
		return
	}
	l.prepare(LOG_INFO, false, format, v...)
}

func SWarn(format string, v ...interface{}) {
	if !isInitLogOnce {
		InitLog("Uninitialized", 1, nil)
	}
	l := GetInstance()
	if LOG_ERROR < l.level {
		return
	}
	l.prepare(LOG_WARN, false, format, v...)
}

func SError(format string, v ...interface{}) {
	if !isInitLogOnce {
		InitLog("Uninitialized", 1, nil)
	}
	l := GetInstance()
	if LOG_ERROR < l.level {
		return
	}
	l.prepare(LOG_ERROR, false, format, v...)
}

func SFatal(format string, v ...interface{}) {
	if !isInitLogOnce {
		InitLog("Uninitialized", 1, nil)
	}
	l := GetInstance()
	if LOG_FATAL < l.level {
		return
	}
	l.prepare(LOG_FATAL, false, format, v...)
}
