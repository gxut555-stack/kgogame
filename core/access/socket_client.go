package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"kgogame/core/access/confuse"
	"kgogame/core/access/ws"
	"kgogame/core/frontProxy/proxy"
	protoAccess "kgogame/proto/access"
	"kgogame/util/logs"
	"kgogame/util/xcodec"
	"net"
	"sync"
	"time"
)

const (
	MSG_DATA_MAX_SIZE      = 1 << 19 //客户端请求消息的最大长度 512k
	MSG_DATA_NORMAL_SIZE   = 1 << 16 //客户端请求消息的默认长度
	READ_TIMEOUT           = 10      //读操作超时(秒)
	WRITE_TIMEOUT          = 10      //写操作超时(秒)
	TIMER_INTERVAL         = 3       //定时周期(秒)
	GUEST_TIMEOUT          = 300     //未登录连接的最大存活时间(秒)
	MAX_IDLE_TIMEOUT       = 60      //客户端最大空闲时间(秒) - 因为有心跳包，如果超过这个时间将被认为连接断开
	PER_SECOND_MAX_REQUEST = 200     //每秒最大请求次数
)

// 客户端链接结构
type SocketClient struct {
	conn          net.Conn            //客户端Socket链接地址
	uid           int32               //登录的UID
	connectAt     int64               //链接时间
	lastActive    int64               //最后活跃时间
	lastReport    int64               //最后上报活跃时间
	addr          string              //客户端地址
	ip            string              //IP
	port          int32               //
	writeChan     chan []byte         //返回数据的通道
	ticker        *time.Ticker        //定时器
	mu            *sync.RWMutex       //锁
	closed        bool                //释放已关闭
	tickerChan    chan int            //定时器上下文
	extData       string              //用于向上游服务透传的扩展数据
	readBuffer    []byte              //读buffer
	beginRead     bool                //是否第一次开始读取
	currentSec    int64               //当前秒数
	requestCnt    int32               //当前秒数请求次数
	isWsShakeHand bool                //是否ws验证握手
	wsHeader      *ws.WsHeader        //ws 头
	mc            *confuse.MsgConfuse //混淆类
	flow          int32               //流水号
	isProxyProto  bool                //是否proxy转发消息过来
}

func NewSocketClient(conn net.Conn, flow int32) *SocketClient {
	ts := time.Now().Unix()
	addr := conn.RemoteAddr().String()

	cli := &SocketClient{
		conn:       conn,
		connectAt:  ts,
		lastActive: ts,
		addr:       addr,
		writeChan:  make(chan []byte, 20),
		ticker:     time.NewTicker(time.Second * TIMER_INTERVAL),
		mu:         new(sync.RWMutex),
		closed:     false,
		tickerChan: make(chan int, 1),
		beginRead:  true,
		readBuffer: make([]byte, MSG_DATA_NORMAL_SIZE),
		currentSec: 0,
		requestCnt: 0,

		mc: &confuse.MsgConfuse{
			RandomArrPos: 6,
			RandomPbPos:  5,
			Rate:         5,
		},
		wsHeader:     &ws.WsHeader{},
		flow:         flow,
		isProxyProto: false,
	}
	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
	remoteIp := remoteAddr.IP.String()
	remotePort := remoteAddr.Port
	cli.ip, cli.port = remoteIp, int32(remotePort)

	//启动协程
	if IsWebSocketServer {
		//go cli.StartRead()
	} else {
		go cli.StartRead()
	}
	go cli.StartTimer()
	return cli
}

// 设置客户端
func (cli *SocketClient) SetClientAddr() {
	if oldCli, exists := Clients.Get(cli.addr); exists {
		oldCli.Close(true, "Start")
	}

	Clients.Set(cli.addr, cli)
}

func (cli *SocketClient) Addr() string {
	return cli.addr
}

func (cli *SocketClient) SetUid(uid int32, extData string) {
	cli.uid = uid
	cli.extData = extData
}

// 关闭客户端一个链接
func (cli *SocketClient) Close(force bool, reason string) {
	defer recoverFunc()

	cli.mu.Lock()
	if cli.closed {
		cli.mu.Unlock()
		logs.Info("socket client %s uid:%d reason:%s already closed", cli.addr, cli.uid, reason)
		return
	}
	cli.closed = true
	if cli.uid > 0 {
		Users.Del(cli.uid, cli.addr)
	}
	Clients.Del(cli.addr)
	close(cli.tickerChan)

	cli.mu.Unlock()
	cli.ticker.Stop()
	//关闭写通道
	close(cli.writeChan)
	//关闭链接
	_ = cli.conn.Close()

	if cli.uid > 0 {
		if force {
			go cli.notifyOffline(cli.uid, cli.ip, cli.port)
		}
	}
}

// 检查是否关闭
func (cli *SocketClient) IsClosed() bool {
	cli.mu.RLock()
	defer cli.mu.RUnlock()
	return cli.closed
}

// 开始读取数据 只读一次
func (cli *SocketClient) BeginRead() error {
	bodySize := uint32(0)
	err := binary.Read(cli.conn, binary.BigEndian, &bodySize)
	if err != nil {
		cli.Close(true, "begin read failed")
		logs.Error("socket client %s begin read failed: %s", cli.addr, err.Error())
		return fmt.Errorf("begin read failed: %s", err.Error())
	}

	//预先读长度
	var preloadBuff []byte
	b1 := make([]byte, 4)
	binary.BigEndian.PutUint32(b1, bodySize)

	//是否是proxy包 再读8位
	b2 := make([]byte, 8)
	if b1[0] == proxy.V2sig[0] && b1[1] == proxy.V2sig[1] && b1[2] == proxy.V2sig[2] && b1[3] == proxy.V2sig[3] {
		if err := binary.Read(cli.conn, binary.BigEndian, &b2); err != nil {
			cli.Close(true, "begin read proxy header failed")
			logs.Error("socket client %s begin read proxy header failed: %s", cli.addr, err.Error())
			return fmt.Errorf("begin read proxy header failed: %s", err.Error())
		}
		//继续读8位
		equalValue := true
		for k, v := range b2 {
			if k+4 == 12 {
				break
			}
			if v != proxy.V2sig[k] {
				equalValue = false
				break
			}
		}
		//比较有异常
		if equalValue {
			preloadBuff = b2
			goto Normal
		}
		//往后只能是proxy 包了
		//再读4位 确认是否是ipV4 ipV6
		b3 := make([]byte, 4)
		if err := binary.Read(cli.conn, binary.BigEndian, &b3); err != nil {
			cli.Close(true, "begin read proxy header failed")
			logs.Error("socket client %s begin read proxy header failed: %s", cli.addr, err.Error())
			return fmt.Errorf("begin read proxy header failed: %s", err.Error())
		}
		fam := b3[0]
		protocalLen := binary.BigEndian.Uint16(b3[2:])
		if fam == 0x11 && protocalLen == 12 { //ipV4
			b4 := make([]byte, 12)
			if err := binary.Read(cli.conn, binary.BigEndian, &b4); err != nil {
				cli.Close(true, "BeginReadd_3")
				logs.Error("ADDR=%s|UID=%d|binary.Read() head BeginReadd_4 error: %s", cli.addr, cli.uid, err.Error())
				return errors.New("read error")
			}

			proxyBuf := make([]byte, len(b1)+len(b2)+len(b3)+len(b4))
			copy(proxyBuf, b1)
			copy(proxyBuf[len(b1):], b2)
			copy(proxyBuf[len(b1)+len(b2):], b3)
			copy(proxyBuf[len(b1)+len(b2)+len(b3):], b4)

			pp, err := proxy.DecodeProxyProtocol(proxyBuf)
			if err != nil {
				cli.Close(true, "BeginReadd_3_1")
				logs.Error("ADDR=%s|UID=%d|binary.Read() head BeginReadd_4_1 error: %s", cli.addr, cli.uid, err.Error())
				return errors.New("DecodeProxyProtocol error")
			}
			//logs.Info("info 4 pp.V4.Src_addr:%d:%d", pp.V4.Src_addr, pp.V4.Src_port)

			cli.ip = proxy.UInt32ToIP(pp.V4.Src_addr).String()
			cli.port = int32(pp.V4.Src_port)

			cli.addr = fmt.Sprintf("%s:%d", cli.ip, cli.port)

			//设置客户端IP
			cli.SetClientAddr()
		} else if fam == 0x21 && protocalLen == 36 {
			//ipv6
			b4 := make([]byte, 36)
			if err := binary.Read(cli.conn, binary.BigEndian, &b4); err != nil {
				cli.Close(true, "BeginReadd_3")
				logs.Error("ADDR=%s|UID=%d|binary.Read() head BeginReadd_5 error: %s", cli.addr, cli.uid, err.Error())
				return errors.New("read error")
			}

			proxyBuf := make([]byte, len(b1)+len(b2)+len(b3)+len(b4))
			copy(proxyBuf, b1)
			copy(proxyBuf[len(b1):], b2)
			copy(proxyBuf[len(b1)+len(b2):], b3)
			copy(proxyBuf[len(b1)+len(b2)+len(b3):], b4)

			pp, err := proxy.DecodeProxyProtocol(proxyBuf)
			if err != nil {
				cli.Close(true, "BeginReadd_3_1")
				logs.Error("ADDR=%s|UID=%d|binary.Read() head BeginReadd_4_1 error: %s", cli.addr, cli.uid, err.Error())
				return errors.New("DecodeProxyProtocol error")
			}
			IpBuff := make([]byte, 16)

			for k, _ := range IpBuff {
				//for protect
				if k > 15 {
					break
				}
				IpBuff[k] = pp.V6.Src_addr[k]
			}
			cli.ip = string(IpBuff)
			cli.port = int32(pp.V6.Src_port)

			cli.addr = fmt.Sprintf("[%s]:%d", cli.ip, cli.port)

			//设置客户端IP
			cli.SetClientAddr()
		} else {
			logs.Error("ADDR=%s|UID=%d|binary.Read() head BeginReadd_4 error: %s", cli.addr, cli.uid)

			preloadBuff = make([]byte, len(b2)+len(b3))
			copy(preloadBuff, b2)
			copy(preloadBuff[len(b2):], b3)

			logs.Info("info 5 2 pp.cli.ip:[%s]:[%d]:[%s]", cli.ip, cli.port, cli.addr)

			goto Normal
		}
		return nil
	}
Normal:
	//长度不正确，主动关闭连接
	if bodySize > MSG_DATA_MAX_SIZE {
		cli.Close(true, "StartRead_2")
		logs.Fatal("ADDR=%s|UID=%d|head size is invalid: %d", cli.addr, cli.uid, bodySize)
		return errors.New("read size error")
	}

	//设置客户端IP
	cli.SetClientAddr()

	logs.Info("info 6 pp.cli.ip:[%s]:[%d]:[%s]", cli.ip, cli.port, cli.addr)

	//读取数据
	var bodyLeftSize uint32
	if preloadBuff == nil || len(preloadBuff) == 0 {
		bodyLeftSize = bodySize
	} else {
		bodyLeftSize = bodySize - uint32(len(preloadBuff))
	}
	logs.Debug("ADDR=%s|UID=%d|bodySize=%d|preloadBuffSize=%d|bodyLeftSize=%d", cli.addr, cli.uid, bodySize, len(preloadBuff), bodyLeftSize)
	if err := cli.ReadBodyAndProcess(preloadBuff, bodyLeftSize); err != nil {
		logs.Fatal("ADDR=%s|UID=%d|ReadAndProcess() error: %s", cli.addr, cli.uid, err.Error())
	}

	return nil
}

func (cli *SocketClient) StartRead() {

	for {
		if cli.IsClosed() {
			break
		}
		//是否首次读
		if cli.beginRead {
			cli.beginRead = false
			if err := cli.BeginRead(); err != nil {
				cli.Close(true, "begin read failed")
				logs.Error("socket client %s begin read failed: %s", cli.addr, err.Error())
				break
			}
		} else {
			//读一个长度
			bodySize := uint32(0)
			err := binary.Read(cli.conn, binary.BigEndian, &bodySize)
			if err != nil {
				cli.Close(true, "read body size failed")
				logs.Error("socket client %s read body size failed: %s", cli.addr, err.Error())
				break
			}
			if bodySize > MSG_DATA_MAX_SIZE {
				cli.Close(true, "read body size overflow")
				logs.Fatal("socket client %s read body size overflow: %d uid:%d", cli.addr, bodySize, cli.uid)
				break
			}
			//读取body
			if err := cli.ReadBodyAndProcess(nil, bodySize); err != nil {
				logs.Fatal("socket client %s read body failed: %s", cli.addr, err.Error())
				break
			}
		}

	}

}

func (cli *SocketClient) ReadBodyAndProcess(preloadBuf []byte, bodyLeftSize uint32) error {
	ts := time.Now().Unix()
	if ts != cli.currentSec {
		cli.currentSec = ts
		cli.requestCnt = 0
	}
	cli.requestCnt++
	if cli.requestCnt >= PER_SECOND_MAX_REQUEST {
		cli.Close(false, "request cnt exceeds max request cnt")
		logs.Error("socket client %s request cnt %d exceeds max request cnt %d", cli.addr, cli.requestCnt, PER_SECOND_MAX_REQUEST)
		return fmt.Errorf("request cnt %d exceeds max request cnt %d", cli.requestCnt, PER_SECOND_MAX_REQUEST)
	}

	//设置读写超时
	if err := cli.conn.SetReadDeadline(time.Now().Add(time.Second * READ_TIMEOUT)); err != nil {
		cli.Close(true, "set read deadline failed")
		logs.Error("socket client %s set read deadline failed: %s", cli.addr, err.Error())
		return fmt.Errorf("set read deadline failed: %s", err.Error())
	}
	//扩容
	if bodyLeftSize+uint32(len(preloadBuf)) > MSG_DATA_MAX_SIZE && len(cli.readBuffer) == MSG_DATA_MAX_SIZE {
		cli.readBuffer = make([]byte, MSG_DATA_NORMAL_SIZE)
	}

	startPos, endPos := 0, 0
	if len(preloadBuf) > 0 {
		{
			for i := 0; i < len(preloadBuf) && i < len(cli.readBuffer); i++ {
				cli.readBuffer[i] = cli.readBuffer[i]
			}
			startPos = len(preloadBuf)
		}
	}
	endPos = startPos + int(bodyLeftSize)
	//读取数据
	if _, err := io.ReadAtLeast(cli.conn, cli.readBuffer[startPos:endPos], int(bodyLeftSize)); err != nil {
		cli.Close(true, "read body failed")
		logs.Error("socket client %s read body failed: %s", cli.addr, err.Error())
		return fmt.Errorf("read body failed: %s", err.Error())

	} else {
		//取消读超时
		_ = cli.conn.SetReadDeadline(time.Time{})

		//解包并处理
		msg := new(xcodec.PacketDataSt)
		if err := msg.UnPack(cli.readBuffer[startPos:endPos]); err != nil {
			cli.Close(true, "unpack packet data failed")
			logs.Error("socket client %s unpack packet data failed: %s", cli.addr, err.Error())
			return fmt.Errorf("unpack packet data failed: %s", err.Error())
		} else {
			//处理消息
			fmt.Printf("Addr=%s Uid=%d Request:#v Cookie:%#v Extend:%#v\n", cli.addr, cli.uid, msg, msg.Cookie, msg.Extend)
			go cli.dispatchRequest(msg)
		}
	}

	return nil
}

// 开始写入数据
func (cli *SocketClient) StartWrite() {
	for {
		data, ok := <-cli.writeChan
		if !ok {
			return
		}
		if IsWebSocketServer {
			//if err := cli.safeWriteWs(data); err != nil {
			//	logs.Error("socket client %s safe write failed: %s", cli.addr, err.Error())
			//	cli.Close(true, "safe write ws failed")
			//	break
			//}
		} else {
			if err := cli.safeWrite(data); err != nil {
				logs.Error("socket client %s safe write failed: %s", cli.addr, err.Error())
				cli.Close(true, "write failed")
				break
			}
		}
	}
}

// 开始定时器循环
func (cli *SocketClient) StartTimer() {
	for {
		select {
		case <-cli.tickerChan:
			return
		case t, ok := <-cli.ticker.C:
			if !ok {
				return
			}
			ts := t.Unix()
			cli.mu.RLock()
			lastActive := cli.lastActive
			cli.mu.RUnlock()
			if ts-lastActive > MAX_IDLE_TIMEOUT {
				cli.Close(false, "idle timeout")
				return
			}
			//检查是否在指定时间内登录
			if cli.uid <= 0 && ts-cli.connectAt > GUEST_TIMEOUT {
				cli.Close(false, "guest timeout")
				return
			}
		}
	}
}

func (cli *SocketClient) safeWrite(data []byte) error {
	defer recoverFunc()

	n := len(data)
	if n == 0 {
		return nil
	}
	//设置写超时
	if err := cli.conn.SetWriteDeadline(time.Now().Add(time.Second * WRITE_TIMEOUT)); err != nil {
		return fmt.Errorf("set write deadline failed: %s", err.Error())
	}

	if cli.IsClosed() {
		return fmt.Errorf("socket client %s is closed", cli.addr)
	}
	//写长度
	if err := binary.Write(cli.conn, binary.BigEndian, uint32(n)); err != nil {
		return fmt.Errorf("write data length failed: %s", err.Error())
	}
	//写入数据
	if writeLen, err := cli.conn.Write(data); err != nil {
		return fmt.Errorf("write data failed: %s", err.Error())
	} else if writeLen != n {
		return fmt.Errorf("write data length mismatch: %d != %d", writeLen, n)
	}
	return nil

}

func (cli *SocketClient) dispatchRequest(msg *xcodec.PacketDataSt) {
	//defer recoverFunc()
	//logs.Debug("Addr=%s Uid=%d Request:#v Cookie:%#v Extend:%#v", cli.addr, cli.uid, msg, msg.Cookie, msg.Extend)
	//
	////更新最后的活跃时间
	//cli.updateLastActive()
	////如果是心跳包直接处理
	//if msg.SvcName == "heat" || msg.FuncName == "heat" {
	//	cli.handleHeatRequest(msg)
	//	return
	//}
	////协议转换
	//req := &protoAccess.RPCRequest{
	//	RemoteIP:        cli.ip,
	//	RemotePort:      cli.port,
	//	AccSrvID:        cli.flow,
	//	Seq:             msg.Seq,
	//	UID:             cli.uid,
	//	Payload:         msg.PayLoad,
	//	ServerName:      msg.SvcName,
	//	FuncName:        msg.FuncName,
	//	RefererServerID: 0,
	//	RefererServer:   "",
	//	ExtData:         cli.extData,
	//}
	//
	//if msg.Cookie != nil {
	//	req.RefererServerID = msg.Cookie.SrvID
	//	req.RefererServer = msg.Cookie.SrvName
	//}
	//ctx, _ := context.WithTimeout(context.Background(), gDefine.TIMEOUT_NORMAL)
	//resp := new(protoAccess.RPCResponse)
	//if err := rpcClient.ProximityIPXCall(ctx, "DispatchService", "Send", req, resp); err != nil {
	//	logs.Fatal("Addr=%s Uid=%d handle ProximityIPXCall(DispatchService,send ,%#v) failed: %s", cli.addr, cli.uid, req, err.Error())
	//	cli.Response(&protoAccess.RPCResponse{Seq: msg.Seq, Code: gDefine.SYS_RESP_CODE_TIME_OUT, Payload: []byte{}})
	//
	//} else if resp != nil && resp.Response {
	//	if len(resp.Payload) > 0 && len(msg.Token) == 1 {
	//		resp.Token = string([]byte{byte(rand.Intn(256))})
	//	}
	//	cli.Response(resp)
	//} else {
	//
	//}
	//
	//if resp != nil && resp.Disconnect {
	//	cli.Close(false, "dispatchRequest")
	//	logs.Error("Addr=%s Uid=%d dispatchRequest disconnect", cli.addr, cli.uid)
	//}
	////熔断处理
	//if resp != nil && resp.Code == gDefine.SYS_FRAME_CODE_CIRCUIT_BREAKER_REQ {
	//	cli.Close(false, "circuit breaker")
	//	logs.Error("Addr=%s Uid=%d dispatchRequest circuit breaker", cli.addr, cli.uid)
	//	return
	//}

}

// 处理心跳包
func (cli *SocketClient) handleHeatRequest(msg *xcodec.PacketDataSt) {
	defer recoverFunc()
	logs.Debug("Addr=%s Uid=%d Request:#v Cookie:%#v Extend:%#v", cli.addr, cli.uid, msg, msg.Cookie, msg.Extend)
	cli.Response(&protoAccess.RPCResponse{
		Seq:       msg.Seq,
		Code:      0,
		Cmd:       0,
		Payload:   []byte{},
		SetCookie: false,
		SrvName:   "",
		SrvID:     0,
	})
}

// 向客户端返回数据(实际上是只把数据写入数据通道)
func (cli *SocketClient) Response(resp *protoAccess.RPCResponse) {
	if IsWebSocketServer {
		return
	}
	defer recoverFunc()

	//返回结构
	rsp := &xcodec.RspDataSt{
		Version: xcodec.VERSION,
		Seq:     resp.Seq,
		Code:    resp.Code,
		Cmd:     resp.Cmd,
		Token:   resp.Token,
		PayLoad: resp.Payload,
	}

	if resp.SetCookie {
		rsp.Cookie = &xcodec.CookieMessage{
			SrvName: resp.SrvName,
			SrvID:   resp.SrvID,
		}
	}

	data, err := rsp.Pack()
	if err != nil {
		logs.Fatal("Addr=%s Uid=%d socket client pack response failed: %s", cli.addr, cli.uid, err.Error())
		return
	}

	if cli.IsClosed() {
		logs.Fatal("Addr=%s Uid=%d socket client is closed", cli.addr, cli.uid)
		return
	}
	logs.Debug("Addr=%s Uid=%d Response:#v Cookie:%#v Extend:%#v cmd:x%.4x", cli.addr, cli.uid, rsp, rsp.Cookie, rsp.Extend, rsp.Cmd)
	//发送数据
	cli.writeChan <- data
}

// 更新最后活跃时间
func (cli *SocketClient) updateLastActive() {
	needReport, ts := false, time.Now().Unix()

	cli.mu.Lock()
	cli.lastActive = ts
	if ts-cli.lastReport > 60 {
		needReport = true
		cli.lastReport = ts
	}
	cli.mu.Unlock()

	if needReport {
		go cli.notifyActive()
	}
}

// 通知online更新最后活跃时间
func (cli *SocketClient) notifyActive() {
	//if cli.uid <= 0 {
	//	return
	//}
	//if _, err := libOnline.Active(nil, cli.uid); err != nil {
	//	logs.Fatal("libOnline.Active(%d) error: %s", cli.uid, err.Error())
	//} else {
	//	logs.Debug("libOnline.Active(%d) success", cli.uid)
	//}
}

// 通知Online客户端断线
func (cli *SocketClient) notifyOffline(uid int32, ip string, port int32) {
	//if uid <= 0 {
	//	return
	//}
	//if _, err := libOnline.Offline(nil, uid, ip, port, gconf.CmdConf.ServerID); err != nil {
	//	logs.Fatal("libOnline.Offline(%d, %s, %d) error: %s", uid, ip, port, err.Error())
	//} else {
	//	logs.Debug("libOnline.Offline(%d, %s, %d) success", uid, ip, port)
	//}
}
