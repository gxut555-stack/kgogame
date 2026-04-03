package logs

import (
	"context"
	"fmt"
)

type Logger struct {
}

// debug级别日志
func (l *Logger) SDebug(format string, v ...interface{}) {
	if !isInitLogOnce {
		InitLog("Uninitialized", 1, nil)
	}
	inst := GetInstance()
	if LOG_DEBUG < inst.level {
		return
	}
	inst.prepare(LOG_DEBUG, false, format, v...)
}

func (l *Logger) Debug(v ...interface{}) {
	if !isInitLogOnce {
		InitLog("Uninitialized", 1, nil)
	}
	inst := GetInstance()
	if LOG_DEBUG < inst.level {
		return
	}
	inst.prepare(LOG_DEBUG, false, "%s", fmt.Sprint(v...))
}

// info级别日志
func (l *Logger) SInfo(format string, v ...interface{}) {
	if !isInitLogOnce {
		InitLog("Uninitialized", 1, nil)
	}
	inst := GetInstance()
	if LOG_INFO < inst.level {
		return
	}
	inst.prepare(LOG_INFO, false, format, v...)
}

func (l *Logger) SWarn(format string, v ...interface{}) {
	if !isInitLogOnce {
		InitLog("Uninitialized", 1, nil)
	}
	inst := GetInstance()
	if LOG_ERROR < inst.level {
		return
	}
	inst.prepare(LOG_ERROR, false, format, v...)
}

func (l *Logger) SError(format string, v ...interface{}) {
	if !isInitLogOnce {
		InitLog("Uninitialized", 1, nil)
	}
	inst := GetInstance()
	if LOG_ERROR < inst.level {
		return
	}
	inst.prepare(LOG_ERROR, false, format, v...)
}

func (l *Logger) SFatal(format string, v ...interface{}) {
	if !isInitLogOnce {
		InitLog("Uninitialized", 1, nil)
	}
	inst := GetInstance()
	if LOG_FATAL < inst.level {
		return
	}
	inst.prepare(LOG_FATAL, false, format, v...)
}

// 输出日志到文件中，DEBUG级别, format为格式化配置
func (l *Logger) CDebug(ctx context.Context, format string, v ...interface{}) {
	if !isInitLogOnce {
		InitLog("Uninitialized", 1, nil)
	}
	inst := GetInstance()
	if LOG_DEBUG < inst.level {
		return
	}

	inst.prepare(LOG_DEBUG, false, genCtxFmt(ctx, format), v...)
}

func (l *Logger) CInfo(ctx context.Context, format string, v ...interface{}) {
	if !isInitLogOnce {
		InitLog("Uninitialized", 1, nil)
	}
	inst := GetInstance()
	if LOG_INFO < inst.level {
		return
	}
	inst.prepare(LOG_INFO, false, genCtxFmt(ctx, format), v...)
}

func (l *Logger) CWarn(ctx context.Context, format string, v ...interface{}) {
	if !isInitLogOnce {
		InitLog("Uninitialized", 1, nil)
	}
	inst := GetInstance()
	if LOG_ERROR < inst.level {
		return
	}
	inst.prepare(LOG_WARN, false, genCtxFmt(ctx, format), v...)
}

func (l *Logger) CError(ctx context.Context, format string, v ...interface{}) {
	if !isInitLogOnce {
		InitLog("Uninitialized", 1, nil)
	}
	inst := GetInstance()
	if LOG_ERROR < inst.level {
		return
	}
	inst.prepare(LOG_ERROR, false, genCtxFmt(ctx, format), v...)
}

func (l *Logger) CFatal(ctx context.Context, format string, v ...interface{}) {
	if !isInitLogOnce {
		InitLog("Uninitialized", 1, nil)
	}
	inst := GetInstance()
	if LOG_FATAL < inst.level {
		return
	}
	inst.prepare(LOG_FATAL, false, genCtxFmt(ctx, format), v...)
}

func NewLogger() *Logger {
	return &Logger{}
}
