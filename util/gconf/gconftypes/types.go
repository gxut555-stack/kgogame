package gconftypes

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	ClientHeartBeatDelay = 15   //定时器的执行间隔
	ClientDataMaxSize    = 1024 //客户端请求消息的最大长度
	ClientReadTimeout    = 10   //读操作超时(秒)
	ClientWriteTimeout   = 10   //写操作超时(秒)
	ClientIdleTimeout    = 60   //客户端最大空闲时间(秒) - 因为有心跳包，如果超过这个时间将被认为连接断开
	ClientMsgHeadSize    = 8    //客户端请求消息最小长度
)

const (
	ServerDataMaxSize = 1 << 19
)

const (
	ClientCmdPing     = 1 //保活
	ClientCmdQuery    = 2 //取数据
	ClientCmdQueryAll = 3 //取全部数据
	ClientCmdPush     = 4 //推送有变化
)

// 请求包
type Msg struct {
	Seq  uint32 //序列号
	Cmd  uint32 //服务名
	Data []byte //透传数据
}

// 编码请求数据
func (msg *Msg) Encode() ([]byte, error) {
	if msg == nil {
		return nil, errors.New("invalid instance")
	}
	buff := bytes.NewBuffer(nil)
	if err := binary.Write(buff, binary.BigEndian, msg.Seq); err != nil {
		return nil, fmt.Errorf("write field seq failed: %s", err.Error())
	}
	if err := binary.Write(buff, binary.BigEndian, msg.Cmd); err != nil {
		return nil, fmt.Errorf("write field cmd failed: %s", err.Error())
	}
	if _, err := buff.Write(msg.Data); err != nil {
		return nil, fmt.Errorf("write field data failed: %s", err.Error())
	} else {
		return buff.Bytes(), nil
	}
}

// 解码请求数据
func (msg *Msg) Decode(data []byte) (*Msg, error) {
	if n := len(data); n < ClientMsgHeadSize {
		return nil, fmt.Errorf("data size is invalid, %d bytes", n)
	}
	msg.Seq = binary.BigEndian.Uint32(data[0:4])
	msg.Cmd = binary.BigEndian.Uint32(data[4:ClientMsgHeadSize])
	msg.Data = data[ClientMsgHeadSize:]
	return msg, nil
}

// 查询配置请求结构
type CmdQueryRequest struct {
	Project string `json:"project"`
	Version int32  `json:"version"`
	Name    string `json:"name"`
}

// 查询配置响应结构
type CmdQueryResponse struct {
	Project string          `json:"project"`
	Version int32           `json:"version"`
	Items   map[string]Item `json:"items"`
	Changed bool            `json:"changed"`
}

// 查询配置响应结构
type CmdQueryAllResponse struct {
	Versions map[string]int32           `json:"versions"`
	Projects map[string]map[string]Item `json:"projects"`
}

// 配置数据记录格式
type Item struct {
	Key    string `json:"key"`    //数据的KEY
	Value  string `json:"value"`  //数据的值
	Format string `json:"format"` //数据的格式
}

// 回调
type GConfUpdateCallback func(project string, version int32, old map[string]Item)
