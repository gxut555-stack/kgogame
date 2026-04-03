package main

import (
	"flag"
	"fmt"
	"kgogame/util/gconf"
	"kgogame/util/logs"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// 全局变量
var (
	SockSrv           *SocketServer //SOCKET服务对象
	Clients           *ClientsMap   //客户端连接的映射表
	Users             *UsersMap     //用户与连接的映射表
	gCfgPath          string
	pprofPort         int
	IsWebSocketServer bool // 是否启动WebSocket服务器
)

func main() {
	flag.StringVar(&gCfgPath, "cfg", "", "global config file path")
	flag.IntVar(&pprofPort, "pprofPort", 6060, "pprof port")
	flag.BoolVar(&IsWebSocketServer, "IsWebSocketServer", false, "是否启动WebSocket服务器")
	if err := gconf.InitConf(); err != nil {
		fmt.Println("init config err ", err)
		return
	}
	if gCfgPath == "" {
		log.Fatal("global config path is empty")
	}

	//初始化日志
	if err := logs.InitializeLog(gconf.CmdConf.ServerName, gconf.CmdConf.ServerID, gCfgPath); err != nil {
		log.Fatalf("logs.InitializeLog(%s) err %s", gCfgPath, err.Error())
		return
	}

	//初始化配置文件
	if err := pConfig.InitConfig(gCfgPath); err != nil {
		log.Fatalf("pConfig InitConfig(%s) err %s", gCfgPath, err.Error())
		return
	}
	//监听信号
	go waitSignal()

	//初始化映射表
	Clients = NewClientsMap()
	Users = NewUserMap()

	//初始化SOCKET服务
	SockSrv = NewSocketServer(fmt.Sprintf(":%d", gconf.CmdConf.AcPort))

	go SockSrv.Start()

	//启动RPC服务
	for {

	}
}

// 捕获panic
func recoverFunc() {
	if err := recover(); err != nil {
		logs.Fatal("recover an error: %v", err)
	}
}

func waitSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGUSR1)
	for {
		sig, ok := <-ch
		if !ok {
			fmt.Printf("signal channel is broken")
			return
		}
		switch sig {
		case syscall.SIGUSR1: //更新配置
			fmt.Printf("Get USR1 signal,reload config")
			if err := pConfig.ReloadConfig(); err != nil {
				fmt.Printf("reload config fail err:%s \n", err.Error())
			}
		}
	}
}

// 停止RPC服务
func stopRPCService() {
	logs.Error("stop RPC service")
	SockSrv.Stop()
	time.Sleep(1 * time.Second)
	//_ = app.Shutdown(context.Background())
}
