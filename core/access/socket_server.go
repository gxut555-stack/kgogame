package main

import (
	"crypto/tls"
	"kgogame/util/logs"
	"log"
	"net"
)

type SocketServer struct {
	add          string
	sock         net.Listener
	cert         *tls.Certificate
	loadCertTime int64 //加载赠书时间
	flow         int32 //流水号
}

func NewSocketServer(addr string) *SocketServer {
	return &SocketServer{
		add:          addr,
		flow:         1,
		loadCertTime: 0,
	}
}

func (ss *SocketServer) Addr() string {
	return ss.add

}

// 启动服务
func (ss *SocketServer) Start() {
	var (
		sock net.Listener
		err  error
	)

	if IsWebSocketServer && pConfig.WebSocket.Tls {
		tlsCfg := &tls.Config{
			GetCertificate: func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
				if ss.cert != nil && pConfig.WsLoadTime <= ss.loadCertTime { //加载时间相同
					return ss.cert, nil
				}
				logs.Debug("serverName:%s remote ip :%s", clientHello.ServerName, clientHello.Conn.RemoteAddr().String())
				// 读取证书和密钥 --此处可以动态更新 , 可以判断加载时间不同
				cert, sErr := tls.LoadX509KeyPair(pConfig.WebSocket.Pem, pConfig.WebSocket.Key)
				if sErr != nil {
					logs.Error("load ssl key fail %s config %#v", sErr.Error(), pConfig.WebSocket)
					return nil, sErr
				}
				ss.cert = &cert
				ss.loadCertTime = pConfig.WsLoadTime
				return ss.cert, nil
			},
		}
		if sock, err = tls.Listen("tcp", ss.add, tlsCfg); err != nil {
			log.Fatalf("tls.Listen(%s) err %s", ss.add, err.Error())
			return
		} else {
			ss.sock = sock
			logs.Info("tls.Listen(%s) success", ss.add)
		}
	} else {
		if sock, err = net.Listen("tcp", ss.add); err != nil {
			log.Fatalf("net.Listen(%s) err %s", ss.add, err.Error())
			return
		} else {
			ss.sock = sock
			logs.Info("net.Listen(%s) success", ss.add)
		}
	}

	//退出关闭socket
	defer ss.sock.Close()

	for {
		if conn, err := ss.sock.Accept(); err != nil {
			log.Fatalf("accept failed: %s", err.Error())
			return
		} else {
			NewSocketClient(conn, ss.GetFlow())
		}
	}
}

func (ss *SocketServer) GetFlow() int32 {
	flow := ss.flow
	ss.flow++
	//int32 最大值
	maxFlow := int32(^uint32(0) >> 1)
	if ss.flow > maxFlow {
		ss.flow = 1
	}
	return flow
}

// 停止服务
func (ss *SocketServer) Stop() {
	ss.sock.Close()
}
