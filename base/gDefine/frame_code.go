package gDefine

// 框架层错误码[20000-21000]
const (
	SYS_FRAME_CODE_LIMIT_REQ           = 20000 //限流
	SYS_FRAME_CODE_CIRCUIT_BREAKER_REQ = 20001 //熔断
)

const (
	CONFIG_CENTER_DATA_NOT_CHANGED = 20100 //数据无变化
)

const (
	ECODE_WEAK_USER_PLAYING_CANOT_ENTER_WEAK = 21000 //用户在游戏中，不能进入弱在线
	ECODE_WEAK_USER_IN_OTHER_WEAK_GAME       = 21001 //用户在游戏其他弱游戏中
	ECODE_WEAK_EXIT_INVALID_GID              = 21002 //退出弱在线时，游戏ID不匹配
)
