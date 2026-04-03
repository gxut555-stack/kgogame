package gDefine

import "time"

const (
	TIMEOUT_SHORT     time.Duration = 3 * time.Second
	TIMEOUT_NORMAL    time.Duration = 10 * time.Second
	TIMEOUT_LONG      time.Duration = 30 * time.Second
	TIMEOUT_GAME_LONG time.Duration = 30 * time.Second
	TIMEOUT_DB_PROXY                = 60 * time.Second //数据库相关超时
	TIMEOUT_SHORTEST  time.Duration = 1 * time.Second
)
