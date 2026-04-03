package gDefine

const (
	SYS_RESP_SUCCESS             = 0 //0表示成功
	SYS_RESP_CODE_ILLEGAL_REQ    = 1 //1表示请求异常(请求参数在acc层就解析不合法)
	SYS_RESP_CODE_TIME_OUT       = 2 //2执行超时
	SYS_RESP_CODE_INTERNAL_ERR   = 3 //服务内部错误比如说拿不到在线状态
	SYS_RESP_CODE_UNSAFE_REQ     = 4 //不安全的请求,eg:未经过合法登录就发过来的请求或者是黑名单中的请求
	SYS_RESP_CODE_CONCURRENT_REQ = 5 //并发的请求,eg:玩家连续的点击报名,并发的
)
