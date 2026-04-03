package main

import "sync"

type ClientsMap struct {
	mgr  map[string]*SocketClient
	lock *sync.RWMutex
}

func NewClientsMap() *ClientsMap {
	return &ClientsMap{
		mgr:  make(map[string]*SocketClient),
		lock: new(sync.RWMutex),
	}

}

func (cm *ClientsMap) Get(addr string) (*SocketClient, bool) {
	cm.lock.RLock()
	defer cm.lock.RUnlock()
	if cli, ok := cm.mgr[addr]; ok {
		return cli, true
	} else {
		return nil, false
	}
}

func (cm *ClientsMap) Set(addr string, cli *SocketClient) (*SocketClient, bool) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	if v, ok := cm.mgr[addr]; ok {
		return v, true
	} else {
		cm.mgr[addr] = cli
		return cli, true
	}
}

func (cm *ClientsMap) Del(addr string) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	delete(cm.mgr, addr)
}
