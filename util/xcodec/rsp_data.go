//编码和解码的地方
//大端 网络字节序

//===================协议说明 Access发给客户端的包格式（对应结构体RspDataSt）
//headerLen		4字节表示后续的字节数（协助定位包大小）不包括自己
//version			4字节
//seq				4字节 序列号（方便客户端做顺序操作）
//code				4字节 状态码
//cmd				4字节 （主要用在服务端主动推消息的时候，客户端通过该命令字才知道该如何解析payLoad）
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

// 响应包
type RspDataSt struct {
	Version int32          //版本号
	Seq     int32          //序列号
	Code    int32          //响应状态码
	Cmd     int32          //命令字
	Token   string         //token值
	PayLoad []byte         //透传数据
	Cookie  *CookieMessage //附加cookie
	Extend  *ExtendMessage //扩展数据
}

// 生成返回包的数据
func (msg *RspDataSt) Pack() ([]byte, error) {
	buff := bytes.NewBuffer(nil)
	//version
	if err := binary.Write(buff, binary.BigEndian, msg.Version); err != nil {
		return nil, err
	}
	//seq
	if err := binary.Write(buff, binary.BigEndian, msg.Seq); err != nil {
		return nil, err
	}
	//code
	if err := binary.Write(buff, binary.BigEndian, msg.Code); err != nil {
		return nil, err
	}
	//cmd
	if err := binary.Write(buff, binary.BigEndian, msg.Cmd); err != nil {
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
	//如果需要加密就先对 payload 加密
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
	//extend
	if msg.Extend != nil {
		//extend pack
		extendData, err := msg.Extend.Pack()
		if err != nil {
			return nil, err
		}
		//extend length
		if err := binary.Write(buff, binary.BigEndian, uint32(len(extendData))); err != nil {
			return nil, err
		}
		//extend
		if _, err := buff.Write(extendData); err != nil {
			return nil, err
		}
	}
	return buff.Bytes(), nil
}

// 解析返回包的数据
func (msg *RspDataSt) UnPack(data []byte) error {
	buff, restLen := bytes.NewReader(data), len(data)
	//version
	if err := binary.Read(buff, binary.BigEndian, &msg.Version); err != nil {
		return fmt.Errorf("read version error: %s", err.Error())
	} else {
		restLen -= 4
	}
	//seq
	if err := binary.Read(buff, binary.BigEndian, &msg.Seq); err != nil {
		return fmt.Errorf("read seq error: %s", err.Error())
	} else {
		restLen -= 4
	}
	//code
	if err := binary.Read(buff, binary.BigEndian, &msg.Code); err != nil {
		return fmt.Errorf("read code error: %s", err.Error())
	} else {
		restLen -= 4
	}
	//cmd
	if err := binary.Read(buff, binary.BigEndian, &msg.Cmd); err != nil {
		return fmt.Errorf("read cmd error: %s", err.Error())
	} else {
		restLen -= 4
	}
	//token
	tokenLen := uint32(0)
	if restLen > 0 {
		if err := binary.Read(buff, binary.BigEndian, &tokenLen); err != nil {
			return fmt.Errorf("read token length error: %s", err.Error())
		} else if tokenLen > MAX_TOKEN_LENGTH {
			return errors.New(fmt.Sprintf("Token length is invalid: %d", tokenLen))
		} else {
			restLen -= 4
			if tokenLen > 0 {
				token := make([]byte, tokenLen)
				if _, err := io.ReadFull(buff, token); err != nil {
					return fmt.Errorf("read token error: %s", err.Error())
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
			return fmt.Errorf("read payload length error: %s", err.Error())
		} else if payloadLen > MAX_PAYLOAD_LENGTH {
			return errors.New(fmt.Sprintf("Payload length is invalid: %d", payloadLen))
		} else {
			restLen -= 4
			if payloadLen > 0 {
				payload := make([]byte, payloadLen)
				if _, err := io.ReadFull(buff, payload); err != nil {
					return fmt.Errorf("read payload error: %s", err.Error())
				} else {
					restLen -= int(payloadLen)
					//判断内容是否有加密，如果加密了，需要将其解密
					if payloadLen > 0 && tokenLen == 1 {
						//解密
						if payloadDec, err := Decrypt(payload, int(msg.Token[0])); err != nil {
							return fmt.Errorf("decrypt token error: %s", err.Error())
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
			return fmt.Errorf("read cookie length error: %s", err.Error())
		} else if cookieLen > MAX_COOKIE_LENGTH {
			return errors.New(fmt.Sprintf("Cookie length is invalid: %d", cookieLen))
		} else {
			restLen -= 4
			if cookieLen > 0 {
				cookieData := make([]byte, cookieLen)
				if _, err := io.ReadFull(buff, cookieData); err != nil {
					return fmt.Errorf("read cookie error: %s", err.Error())
				} else {
					restLen -= int(cookieLen)
					cookie := new(CookieMessage)
					if err = cookie.UnPack(cookieData); err != nil {
						return fmt.Errorf("unpack cookie error: %s", err.Error())
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
			return fmt.Errorf("read extend data length error: %s", err.Error())
		} else if extendLen > MAX_EXTEND_LENGTH {
			return errors.New(fmt.Sprintf("Extend length is invalid: %d", extendLen))
		} else {
			restLen -= 4
			if extendLen > 0 {
				extendData := make([]byte, extendLen)
				if _, err := io.ReadFull(buff, extendData); err != nil {
					return fmt.Errorf("read extend data error: %s", err.Error())
				} else {
					restLen -= int(extendLen)
					extend := new(ExtendMessage)
					if err = extend.UnPack(extendData); err != nil {
						return fmt.Errorf("unpack extend data error: %s", err.Error())
					} else {
						msg.Extend = extend
					}
				}
			}
		}
	}
	return nil
}

// 响应包的String方法！方便打印查看（fmt中打印时会选择String返回的数据打印）
func (msg *RspDataSt) String() string {
	return fmt.Sprintf(`RspDataSt{
		Version: %d
		Seq:%d
		Code:%d
		Cmd：%d
		Token:%s
		PayLoad:%x
	}`, msg.Version, msg.Seq, msg.Code, msg.Cmd, msg.Token, msg.PayLoad)
}
