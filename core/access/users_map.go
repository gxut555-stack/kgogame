package main

import "sync"

// 用户ID与客户端地址映射表
type UsersMap struct {
	users  map[int32]string
	locker *sync.RWMutex
}

func NewUserMap() *UsersMap {
	return &UsersMap{
		users:  make(map[int32]string),
		locker: new(sync.RWMutex),
	}
}

// 根据UID读取对应客户端地址
func (um *UsersMap) Get(uid int32) (string, bool) {
	um.locker.RLocker()
	defer um.locker.RUnlock()
	if v, ok := um.users[uid]; ok {
		return v, true
	} else {
		return "", false
	}
}

// 设置UID与客户端地址的映射关系
func (um *UsersMap) Set(uid int32, addr string, overwrite bool) (string, bool) {
	um.locker.Lock()
	defer um.locker.Unlock()
	if v, ok := um.users[uid]; ok && !overwrite {
		return v, false
	} else {
		um.users[uid] = addr
		return addr, true
	}
}

// 删除
func (um *UsersMap) Del(uid int32, addr string) (string, bool) {
	um.locker.Lock()
	defer um.locker.Unlock()
	if v, ok := um.users[uid]; ok && v == addr {
		delete(um.users, uid)
		return v, true
	} else {
		return v, false
	}
}
