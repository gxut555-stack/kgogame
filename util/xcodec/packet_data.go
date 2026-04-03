//编码和解码的地方
//大端 网络字节序

//===================协议说明 客户端发给Access的包的格式
//headerLen		4字节表示后续的字节数（协助定位包大小）不包括自己
//version			4字节
//seq				4字节 序列号（方便客户端做顺序操作）
//serviceNameLen	4字节（用来表示后续的服务名的字节长度）
//serviceName		N字节（通过serviceNameLen来确定N值）【异或混淆】
//funcNameLen		4字节（用来表示方法名称的长度）
//funcName		N字节（由funcNameLen来决定）【异或混淆】
//tokenLen		4字节（用来表示token的长度）
//token			N字节（由tokenLen决定）
//payLoadLen		4字节（业务pb包的大小）
//payLoad			N字节（pb数据包）

package xcodec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// 请求包
type PacketDataSt struct {
	Version  int32          //版本号
	Seq      int32          //序列号
	SvcName  string         //服务名
	FuncName string         //方法名
	Token    string         //token值
	PayLoad  []byte         //透传数据
	Cookie   *CookieMessage //附加数据
	Extend   *ExtendMessage //扩展数据
}

// 生成请求包的数据
func (msg *PacketDataSt) Pack() ([]byte, error) {
	buff := bytes.NewBuffer(nil)
	//version
	if err := binary.Write(buff, binary.BigEndian, msg.Version); err != nil {
		return nil, err
	}
	//seq
	if err := binary.Write(buff, binary.BigEndian, msg.Seq); err != nil {
		return nil, err
	}
	//svcName length
	if err := binary.Write(buff, binary.BigEndian, uint32(len(msg.SvcName))); err != nil {
		return nil, err
	}
	//svcName
	if _, err := buff.WriteString(msg.SvcName); err != nil {
		return nil, err
	}
	//funcName length
	if err := binary.Write(buff, binary.BigEndian, uint32(len(msg.FuncName))); err != nil {
		return nil, err
	}
	//funcName
	if _, err := buff.WriteString(msg.FuncName); err != nil {
		return nil, err
	}
	//token length
	tokenLen := len(msg.Token)
	if err := binary.Write(buff, binary.BigEndian, uint32(tokenLen)); err != nil {
		return nil, err
	}
	//token
	if tokenLen > 0 {
		if _, err := buff.WriteString(msg.Token); err != nil {
			return nil, err
		}
	}
	//如果需要对 payload 加密就先加密
	if msg.PayLoad == nil {
		msg.PayLoad = []byte{}
	}
	if len(msg.PayLoad) > 0 && tokenLen == 1 {
		msg.PayLoad = Encrypt(msg.PayLoad, int(msg.Token[0]))
	}
	//payload length
	payloadLen := len(msg.PayLoad)
	if err := binary.Write(buff, binary.BigEndian, uint32(payloadLen)); err != nil {
		return nil, err
	}
	//payload
	if payloadLen > 0 {
		if _, err := buff.Write(msg.PayLoad); err != nil {
			return nil, err
		}
	}
	//cookie
	if msg.Cookie != nil {
		//cookie pack
		cookieData, err := msg.Cookie.Pack()
		if err != nil {
			return nil, err
		}
		//cookie length
		if err := binary.Write(buff, binary.BigEndian, uint32(len(cookieData))); err != nil {
			return nil, err
		}
		//cookie
		if _, err := buff.Write(cookieData); err != nil {
			return nil, err
		}
	}
	//Extend
	if msg.Extend != nil {
		//Extend pack
		extendData, err := msg.Extend.Pack()
		if err != nil {
			return nil, err
		}
		//Extend length
		if err := binary.Write(buff, binary.BigEndian, uint32(len(extendData))); err != nil {
			return nil, err
		}
		//Extend
		if _, err := buff.Write(extendData); err != nil {
			return nil, err
		}
	}
	return buff.Bytes(), nil
}

// 解析请求包的数据
func (msg *PacketDataSt) UnPack(data []byte) error {
	buff, restLen := bytes.NewReader(data), len(data)
	//version
	if err := binary.Read(buff, binary.BigEndian, &msg.Version); err != nil {
		return err
	} else {
		restLen -= 4
	}
	//seq
	if err := binary.Read(buff, binary.BigEndian, &msg.Seq); err != nil {
		return err
	} else {
		restLen -= 4
	}
	//svcName
	svcNameLen := uint32(0)
	if err := binary.Read(buff, binary.BigEndian, &svcNameLen); err != nil {
		return err
	} else if svcNameLen > MAX_SERVICE_LENGTH {
		return errors.New(fmt.Sprintf("SvcName length is invalid: %d", svcNameLen))
	} else {
		restLen -= 4
		if svcNameLen > 0 {
			svcName := make([]byte, svcNameLen)
			if _, err = io.ReadFull(buff, svcName); err != nil {
				return err
			} else {
				msg.SvcName = string(svcName)
				restLen -= int(svcNameLen)
			}
		}
	}
	//funcName
	funcNameLen := uint32(0)
	if err := binary.Read(buff, binary.BigEndian, &funcNameLen); err != nil {
		return err
	} else if funcNameLen > MAX_FUNCTION_LENGTH {
		return errors.New(fmt.Sprintf("FuncName length is invalid: %d", funcNameLen))
	} else {
		restLen -= 4
		if funcNameLen > 0 {
			funcName := make([]byte, funcNameLen)
			if _, err := io.ReadFull(buff, funcName); err != nil {
				return err
			} else {
				msg.FuncName = string(funcName)
				restLen -= int(funcNameLen)
			}
		}
	}
	//token
	tokenLen := uint32(0)
	if restLen > 0 {
		if err := binary.Read(buff, binary.BigEndian, &tokenLen); err != nil {
			return err
		} else if tokenLen > MAX_TOKEN_LENGTH {
			return errors.New(fmt.Sprintf("Token length is invalid: %d", tokenLen))
		} else {
			restLen -= 4
			if tokenLen > 0 {
				token := make([]byte, tokenLen)
				if _, err := io.ReadFull(buff, token); err != nil {
					return err
				} else {
					msg.Token = string(token)
					restLen -= int(tokenLen)
				}
			}
		}
	}
	//payload
	if restLen > 0 {
		payloadLen := uint32(0)
		if err := binary.Read(buff, binary.BigEndian, &payloadLen); err != nil {
			return err
		} else if payloadLen > MAX_PAYLOAD_LENGTH {
			return errors.New(fmt.Sprintf("Payload length is invalid: %d", payloadLen))
		} else {
			restLen -= 4
			if payloadLen > 0 {
				payload := make([]byte, payloadLen)
				if _, err := io.ReadFull(buff, payload); err != nil {
					return err
				} else {
					restLen -= int(payloadLen)
					//判断内容是否有加密，如果加密了，需要将其解密
					if payloadLen > 0 && tokenLen == 1 {
						//解密
						if payloadDec, err := Decrypt(payload, int(msg.Token[0])); err != nil {
							return err
						} else {
							msg.PayLoad = payloadDec
						}
					} else {
						msg.PayLoad = payload
					}
				}
			}
		}
	}
	//cookie
	if restLen > 0 {
		cookieLen := uint32(0)
		if err := binary.Read(buff, binary.BigEndian, &cookieLen); err != nil {
			return err
		} else if cookieLen > MAX_COOKIE_LENGTH {
			return errors.New(fmt.Sprintf("Cookie length is invalid: %d", cookieLen))
		} else {
			restLen -= 4
			if cookieLen > 0 {
				cookieData := make([]byte, cookieLen)
				if _, err := io.ReadFull(buff, cookieData); err != nil {
					return err
				} else {
					restLen -= int(cookieLen)
					cookie := new(CookieMessage)
					if err = cookie.UnPack(cookieData); err != nil {
						return err
					} else {
						msg.Cookie = cookie
					}
				}
			}
		}
	}
	//extend
	if restLen > 0 {
		extendLen := uint32(0)
		if err := binary.Read(buff, binary.BigEndian, &extendLen); err != nil {
			return err
		} else if extendLen > MAX_EXTEND_LENGTH {
			return errors.New(fmt.Sprintf("Extend length is invalid: %d", extendLen))
		} else {
			restLen -= 4
			if extendLen > 0 {
				extendData := make([]byte, extendLen)
				if _, err := io.ReadFull(buff, extendData); err != nil {
					return err
				} else {
					restLen -= int(extendLen)
					extend := new(ExtendMessage)
					if err = extend.UnPack(extendData); err != nil {
						return err
					} else {
						msg.Extend = extend
					}
				}
			}
		}
	}
	return nil
}

// 方便打印查看
func (msg *PacketDataSt) String() string {
	return fmt.Sprintf(`PacketDataSt{
		Version: %d
		Seq:%d
		SvcName：%s
		FuncName:%s
		Token:%s
		PayLoad:%x
	}`, msg.Version, msg.Seq, msg.SvcName, msg.FuncName, msg.Token, msg.PayLoad)
}
