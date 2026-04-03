package xcodec

//常量
const (
	VERSION = 1 //协议版本号
)

//常量
const (
	MAX_SERVICE_LENGTH  = 50      //RPC服务名称最大长度
	MAX_FUNCTION_LENGTH = 50      //RPC方法名最大长度
	MAX_TOKEN_LENGTH    = 200     //TOKEN字段最大长度
	MAX_PAYLOAD_LENGTH  = 1 << 19 //透传数据最大长度
	MAX_COOKIE_LENGTH   = 500     //COOKIE最大长度
	MAX_EXTEND_LENGTH   = 500     //Extend最大长度
)