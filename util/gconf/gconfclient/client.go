package gconfclient

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"kgogame/util/gconf/gconftypes"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// 客户端结构
type GConfClient struct {
	name      string
	addr      string
	project   string
	connected bool
	conn      net.Conn
	seq       uint32
	ch        chan *gconftypes.Msg
	locker    *sync.RWMutex
	started   bool
	startChan chan int
	//---- 来自服务器的缓存数据 ----
	caches   map[string]map[string]gconftypes.Item
	versions map[string]int32
	cb       gconftypes.GConfUpdateCallback
}

// 新实例
func NewGConfClient() *GConfClient {
	name := filepath.Base(os.Args[0])
	cli := &GConfClient{
		name: name,
		//addr:      addr,
		//project:   project,
		ch:        make(chan *gconftypes.Msg, 20),
		locker:    new(sync.RWMutex),
		startChan: make(chan int, 1),
		caches:    make(map[string]map[string]gconftypes.Item),
		versions:  make(map[string]int32),
		cb:        nil,
	}
	//if project != "*" {
	//	cli.versions[project] = version
	//}
	return cli
}

// 判断是否是从服模式
func (cli *GConfClient) isCluster() bool {
	return cli.project == "*"
}

// 自增并返回请求序列号
func (cli *GConfClient) incSeq() uint32 {
	cli.locker.Lock()
	defer cli.locker.Unlock()
	cli.seq++
	return cli.seq
}

// 取项目配置的版本号
func (cli *GConfClient) Version(project string) int32 {
	cli.locker.RLock()
	defer cli.locker.RUnlock()
	return cli.versions[project]
}

// 连接服务器
func (cli *GConfClient) connect() error {
	//判断是否有连接
	if cli.isConnected() {
		return nil
	}
	//建立新连接
	if conn, err := net.Dial("tcp", cli.addr); err != nil {
		return fmt.Errorf("dial %s failed: %s", cli.addr, err.Error())
	} else {
		saved := false
		//再次判断是否连接过了，如果没有就保存连接
		cli.locker.Lock()
		if !cli.connected {
			cli.conn, cli.connected, saved = conn, true, true
		}
		cli.locker.Unlock()
		//如果没有使用新的连接，关闭未没用的连接
		if !saved {
			_ = conn.Close()
		}
		return nil
	}
}

// 重新连接服务器
func (cli *GConfClient) reconnect() error {
	cli.close()
	if err := cli.connect(); err != nil {
		return err
	}
	if err := cli.queryServer(""); err != nil {
		return err
	}
	return nil
}

// 关闭连接
func (cli *GConfClient) close() {
	cli.locker.Lock()
	defer cli.locker.Unlock()
	if cli.connected {
		cli.connected = false
		_ = cli.conn.Close()
	}
}

// 判断是否已连接
func (cli *GConfClient) isConnected() bool {
	cli.locker.RLock()
	defer cli.locker.RUnlock()
	return cli.connected
}

// 判断是否已启动
func (cli *GConfClient) isStarted() bool {
	cli.locker.RLock()
	defer cli.locker.RUnlock()
	return cli.started
}

// 从服务器读一条消息
func (cli *GConfClient) readMsg() (*gconftypes.Msg, error) {
	if !cli.isConnected() {
		return nil, errors.New("connection is closed")
	}
	dataSize := uint32(0)
	if err := binary.Read(cli.conn, binary.BigEndian, &dataSize); err != nil {
		return nil, fmt.Errorf("binary.Read() read size error: %s", err.Error())
	}
	if dataSize > gconftypes.ServerDataMaxSize {
		return nil, fmt.Errorf("size is invalid: %d", dataSize)
	}
	data := make([]byte, dataSize)
	if _, err := io.ReadFull(cli.conn, data); err != nil {
		return nil, fmt.Errorf("binary.Read() read data error: %s", err.Error())
	}
	msg := new(gconftypes.Msg)
	return msg.Decode(data)
}

// 向服务器发送一条消息
func (cli *GConfClient) writeMsg(msg *gconftypes.Msg) error {
	defer recoverFunc()
	if cli.isConnected() {
		timer := time.NewTimer(time.Second * gconftypes.ClientWriteTimeout)
		select {
		case <-timer.C:
			cli.close()
			return errors.New("write data into channel timeout")
		case cli.ch <- msg:
			timer.Stop()
			return nil
		}
	} else {
		return errors.New("connection is closed")
	}
}

// 更新配置
func (cli *GConfClient) updateCaches(data map[string]map[string]gconftypes.Item, versions map[string]int32) {
	cli.locker.Lock()
	started := cli.started
	oldCaches := cli.caches
	cli.caches, cli.versions = data, versions
	if !started {
		cli.started = true
	}
	cli.locker.Unlock()

	if started { //启动后数据有更新调用回调通知业务方
		if cli.cb != nil {
			for k, _ := range data {
				cli.cb(k, versions[k], oldCaches[k])
			}
		}
	} else { //未启动时更新数据后设置启动完成状态
		cli.startChan <- 1
	}

	str, _ := json.Marshal(data)
	fmt.Printf("[gconfclient.readLoop] update config items: %s\n", string(str))
	str, _ = json.Marshal(versions)
	fmt.Printf("[gconfclient.readLoop] update config versions: %s\n", string(str))
}

// 读循环
func (cli *GConfClient) readLoop() {
	for {
		if msg, err := cli.readMsg(); err != nil {
			fmt.Printf("[gconfclient.readLoop] read failed: %s\n", err.Error())
			if err = cli.reconnect(); err != nil {
				fmt.Printf("[gconfclient.readLoop] reconnect failed: %s\n", err.Error())
				time.Sleep(time.Second * gconftypes.ClientHeartBeatDelay)
			}
			continue
		} else {
			switch msg.Cmd {
			case gconftypes.ClientCmdPing: //来自服务器的心跳回包

			case gconftypes.ClientCmdQuery: //查询某个项目配置的响应包
				resp := new(gconftypes.CmdQueryResponse)
				if err = json.Unmarshal(msg.Data, resp); err != nil {
					fmt.Printf("[gconfclient.readLoop] json unmarshal '%s' error: %s\n", string(msg.Data), err.Error())
				} else if resp.Changed && cli.Version(resp.Project) < resp.Version { //当本地配置数据版本号低于返回数据时，更新配置数据
					projects, versions := make(map[string]map[string]gconftypes.Item), make(map[string]int32)
					projects[resp.Project], versions[resp.Project] = resp.Items, resp.Version
					cli.updateCaches(projects, versions)
				}

			case gconftypes.ClientCmdQueryAll: //查询所有项目配置的响应包
				resp := new(gconftypes.CmdQueryAllResponse)
				if err = json.Unmarshal(msg.Data, resp); err != nil {
					fmt.Printf("[gconfclient.readLoop] json unmarshal '%s' error: %s\n", string(msg.Data), err.Error())
				} else {
					cli.updateCaches(resp.Projects, resp.Versions)
				}

			case gconftypes.ClientCmdPush: //服务器通知某个项目的配置有更新
				resp := new(gconftypes.CmdQueryRequest)
				if err = json.Unmarshal(msg.Data, resp); err != nil {
					fmt.Printf("[gconfclient.readLoop] json unmarshal '%s' error: %s\n", string(msg.Data), err.Error())
				} else {
					if err := cli.queryServer(resp.Project); err != nil {
						fmt.Printf("[gconfclient.readLoop] query server data error: %s\n", err.Error())
					}
				}
			}
		}
	}
}

// 写循环
func (cli *GConfClient) writeLoop() {
	for {
		msg, ok := <-cli.ch
		if !ok {
			fmt.Printf("[gconfclient.writeLoop] channel is closed\n")
			break
		}
		//数据编码
		data, err := msg.Encode()
		if err != nil {
			fmt.Println(err.Error())
		} else {
			//写数据
			if err = binary.Write(cli.conn, binary.BigEndian, uint32(len(data))); err != nil {
				fmt.Printf("[gconfclient.writeLoop] binary.Write() write data size error: %s\n", err.Error())
			} else if _, err = cli.conn.Write(data); err != nil {
				fmt.Printf("[gconfclient.writeLoop] binary.Write() write data error: %s\n", err.Error())
			}
			//判断状态
			if err != nil {
				if err = cli.reconnect(); err != nil {
					fmt.Printf("[gconfclient.writeLoop] reconnect failed: %s\n", err.Error())
				}
			}
		}
	}
}

// 心跳循环
func (cli *GConfClient) heartBeat() {
	msg := &gconftypes.Msg{Cmd: gconftypes.ClientCmdPing}
	for {
		time.Sleep(time.Second * gconftypes.ClientHeartBeatDelay)
		msg.Seq = cli.incSeq()
		if err := cli.writeMsg(msg); err != nil {
			if err = cli.reconnect(); err != nil {
				fmt.Printf("[gconfclient.heartBeat] reconnect failed: %s\n", err.Error())
			}
		}
	}
}

// 查询服务器
func (cli *GConfClient) queryServer(project string) error {
	req := gconftypes.CmdQueryRequest{Name: cli.name}
	if project == "" {
		req.Project = cli.project
	} else {
		req.Project = project
	}
	if req.Project != "*" {
		req.Version = cli.Version(req.Project)
	}
	msg := &gconftypes.Msg{Seq: cli.incSeq(), Cmd: gconftypes.ClientCmdQuery}
	if req.Project == "*" {
		msg.Cmd = gconftypes.ClientCmdQueryAll
	}
	msg.Data, _ = json.Marshal(req)
	if err := cli.writeMsg(msg); err != nil {
		return err
	} else {
		return nil
	}
}

// 开始工作
func (cli *GConfClient) Start(addr string, project string, version int32, cb gconftypes.GConfUpdateCallback) error {
	if cli.isStarted() {
		return errors.New("started already")
	}

	cli.locker.Lock()
	if cli.addr == "" && addr != "" {
		cli.addr = addr
	}
	if cli.project == "" && project != "" {
		cli.project = project
		if project != "*" {
			cli.versions[project] = version
		}
	}
	if cli.cb == nil && cb != nil {
		cli.cb = cb
	}
	cli.locker.Unlock()

	if err := cli.connect(); err != nil {
		return err
	}
	if err := cli.queryServer(""); err != nil {
		return err
	}
	go cli.readLoop()
	go cli.writeLoop()
	go cli.heartBeat()
	//等待数据返回
	timer := time.NewTimer(time.Second * gconftypes.ClientReadTimeout)
	select {
	case <-cli.startChan:
		timer.Stop()
		return nil
	case <-timer.C:
		return errors.New("load data from server timeout")
	}
}

// 取单个配置值
func (cli *GConfClient) Item(key string) (gconftypes.Item, bool) {
	if !cli.isStarted() || cli.isCluster() {
		return gconftypes.Item{}, false
	}
	cli.locker.RLock()
	defer cli.locker.RUnlock()
	if items, ok := cli.caches[cli.project]; ok {
		if item, ok := items[key]; ok {
			return item, true
		}
	}
	return gconftypes.Item{}, false
}

// 取指定项目的配置列表
func (cli *GConfClient) ProjectItems(project string) (map[string]gconftypes.Item, bool) {
	ret := make(map[string]gconftypes.Item)
	if !cli.isStarted() {
		return ret, false
	}
	cli.locker.RLock()
	defer cli.locker.RUnlock()
	if project == "" && cli.project != "*" {
		project = cli.project
	}
	items, ok := cli.caches[project]
	if !ok || items == nil {
		return ret, false
	}
	for k, v := range items {
		ret[k] = v
	}
	return ret, true
}

// 取所有项目的配置列表
func (cli *GConfClient) AllProjectsItems() map[string]map[string]gconftypes.Item {
	ret := make(map[string]map[string]gconftypes.Item)
	if !cli.isStarted() {
		return ret
	}
	cli.locker.RLock()
	defer cli.locker.RUnlock()
	for k, v := range cli.caches {
		if v == nil {
			continue
		}
		items := make(map[string]gconftypes.Item)
		for kk, vv := range v {
			items[kk] = vv
		}
		ret[k] = items
	}
	return ret
}

// 取所有项目的版本号
func (cli *GConfClient) AllProjectsVersions() map[string]int32 {
	ret := make(map[string]int32)
	cli.locker.RLock()
	defer cli.locker.RUnlock()
	for k, v := range cli.versions {
		ret[k] = v
	}
	return ret
}

// 取所有项目的KEY
func (cli *GConfClient) AllProjectKeys() []string {
	ret := make([]string, 0)
	cli.locker.RLock()
	defer cli.locker.RUnlock()
	for k, _ := range cli.caches {
		ret = append(ret, k)
	}
	return ret
}

// 捕获panic
func recoverFunc() {
	if err := recover(); err != nil {
		fmt.Printf("[gconfclient.recoverFunc] recover an error: %v\n", err)
	}
}
