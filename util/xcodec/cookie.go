package xcodec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const MAX_COOKIE_SERVER_NAME_LENGTH = 40     //COOKIE中的服务名最大长度

//COOKIE数据
type CookieMessage struct {
	SrvName string //服务名
	SrvID   int32  //服务ID
}

//打包
func (cookie *CookieMessage) Pack() ([]byte, error) {
	if cookie == nil {
		return []byte{}, nil
	}
	if cookie.SrvName == "" || cookie.SrvID <= 0 {
		return []byte{}, nil
	}
	buff := bytes.NewBuffer(nil)
	if err := binary.Write(buff, binary.BigEndian, uint32(len(cookie.SrvName))); err != nil {
		return nil, err
	}
	if _, err := buff.WriteString(cookie.SrvName); err != nil {
		return nil, err
	}
	if err := binary.Write(buff, binary.BigEndian, cookie.SrvID); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

//解包
func (cookie *CookieMessage) UnPack(data []byte) error {
	if data == nil || len(data) == 0 {
		return nil
	}
	buff := bytes.NewReader(data)
	//serverName
	serverNameLen := uint32(0)
	if err := binary.Read(buff, binary.BigEndian, &serverNameLen); err != nil {
		return err
	} else if serverNameLen > MAX_COOKIE_SERVER_NAME_LENGTH {
		return errors.New(fmt.Sprintf("Cookie server name length is invalid: %d", serverNameLen))
	} else {
		serverName := make([]byte, serverNameLen)
		if _, err = io.ReadFull(buff, serverName); err != nil {
			return err
		} else {
			cookie.SrvName = string(serverName)
		}
	}
	//serverID
	if err := binary.Read(buff, binary.BigEndian, &cookie.SrvID); err != nil {
		return err
	}
	return nil
}
