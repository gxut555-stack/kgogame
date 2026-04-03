package xcodec

import (
	"bytes"
	"encoding/binary"
	"io"
)

//扩展信息---支持多子游戏协议
type ExtendMessage struct {
	SrvID   int32  //服务ID
	TableId int64  //桌子ID
	Extend1 int64  //ClubId
	Extend2 int64  //MatchId
	SrvName string //服务名
	PayLoad []byte //扩展字节流 -- 未使用
}

//打包
func (extend *ExtendMessage) Pack() ([]byte, error) {
	if extend == nil {
		return []byte{}, nil
	}
	if extend.SrvName == "" || extend.SrvID <= 0 {
		return []byte{}, nil
	}
	buff := bytes.NewBuffer(nil)
	//服务id
	if err := binary.Write(buff, binary.BigEndian, extend.SrvID); err != nil {
		return nil, err
	}
	if err := binary.Write(buff, binary.BigEndian, extend.TableId); err != nil {
		return nil, err
	}
	//扩展字段1
	if err := binary.Write(buff, binary.BigEndian, extend.Extend1); err != nil {
		return nil, err
	}
	if err := binary.Write(buff, binary.BigEndian, extend.Extend2); err != nil {
		return nil, err
	}
	//服务名
	if err := binary.Write(buff, binary.BigEndian, uint32(len(extend.SrvName))); err != nil {
		return nil, err
	}
	if _, err := buff.WriteString(extend.SrvName); err != nil {
		return nil, err
	}
	//payload length
	if extend.PayLoad == nil {
		extend.PayLoad = []byte{}
	}
	if err := binary.Write(buff, binary.BigEndian, uint32(len(extend.PayLoad))); err != nil {
		return nil, err
	}
	//payload
	if _, err := buff.Write(extend.PayLoad); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

//解包
func (extend *ExtendMessage) UnPack(data []byte) error {
	if data == nil || len(data) == 0 {
		return nil
	}
	buff := bytes.NewReader(data)
	//serverID
	if err := binary.Read(buff, binary.BigEndian, &extend.SrvID); err != nil {
		return err
	}
	//桌子ID
	if err := binary.Read(buff, binary.BigEndian, &extend.TableId); err != nil {
		return err
	}
	if err := binary.Read(buff, binary.BigEndian, &extend.Extend1); err != nil {
		return err
	}
	if err := binary.Read(buff, binary.BigEndian, &extend.Extend2); err != nil {
		return err
	}
	//serverName
	serverNameLen := uint32(0)
	if err := binary.Read(buff, binary.BigEndian, &serverNameLen); err != nil {
		return err
	} else {
		serverName := make([]byte, serverNameLen)
		if _, err = io.ReadFull(buff, serverName); err != nil {
			return err
		} else {
			extend.SrvName = string(serverName)
		}
	}
	payloadLen := uint32(0)
	if err := binary.Read(buff, binary.BigEndian, &payloadLen); err != nil {
		return err
	}
	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(buff, payload); err != nil {
		return err
	} else {
		extend.PayLoad = payload
	}
	return nil
}
