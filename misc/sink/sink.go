package sink

import (
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"kgogame/util/gredis"
	"kgogame/util/logs"
	"sync"
	"time"
)

type (
	CacheData struct {
		Key  string
		Data any
	}

	GetCacheDataFn = func(key string) (*CacheData, error) // 获取缓存数据
	SaveData2DBFn  = func(list []*CacheData) error        // 保存数据到DB
)

func NewSinkLogic(name, rds string, tkDuration time.Duration, watch string, batch int32, getFn GetCacheDataFn, saveFn SaveData2DBFn) *SinkLogic {
	return &SinkLogic{
		name:         name,
		rds:          rds,
		tickDuration: tkDuration,
		watchKey:     watch,
		batch:        batch,
		getCacheFn:   getFn,
		save2DBFn:    saveFn,
	}
}

type SinkLogic struct {
	name         string        // 类型
	rds          string        // redis
	tickDuration time.Duration // 间隔
	watchKey     string        // 监听KEY
	batch        int32         // 批量
	getCacheFn   GetCacheDataFn
	save2DBFn    SaveData2DBFn
}

func (lgc *SinkLogic) keyWatch() string {
	return lgc.watchKey
}

func (lgc *SinkLogic) AddWatch(member string) {
	key := lgc.keyWatch()
	_, err := gredis.Call(lgc.rds, "SADD", key, member)
	if err != nil {
		logs.SFatal("Redis %s SADD %s %s failed %s", lgc.rds, key, member, err)
	}
}

func (lgc *SinkLogic) countWatch() int64 {
	key := lgc.keyWatch()
	val, err := redis.Int64(gredis.Call(lgc.rds, "SCARD", key))
	if err != nil {
		logs.SFatal("Redis %s SCARD %s failed %s", lgc.rds, key, err)
	}
	return val
}

func (lgc *SinkLogic) spopWatchKeys(num int32) ([]string, error) {
	key := lgc.keyWatch()
	sl, err := redis.Strings(gredis.Call(lgc.rds, "SPOP", key, num))
	if errors.Is(err, redis.ErrNil) {
		return nil, nil
	}
	return sl, err
}

// 启动定时
func (lgc *SinkLogic) ProcTick(wg *sync.WaitGroup, done <-chan struct{}) {
	logs.SInfo("%s run", lgc.name)
	defer wg.Done()

	tk := time.NewTicker(lgc.tickDuration)
	defer tk.Stop()

	for {
		select {
		case <-tk.C:
			totalCount := lgc.countWatch()
			success, err := lgc.procSync2DB(lgc.batch)
			if err != nil {
				logs.SFatal("sync2DB failed %s", err)
			}
			if totalCount > 0 {
				logs.SInfo("%s totalCount %d success %d", lgc.name, totalCount, success)
			}

		case <-done:
			fmt.Printf("ProcTick name %s DONE\n", lgc.name)
			return
		}
	}
}

// 进入同步
func (lgc *SinkLogic) procSync2DB(num int32) (succ int, err error) {
	defer logs.CatchPanic()

	for {
		var n int
		n, err = lgc.sync2DB(num)
		if err != nil {
			return succ, err
		}
		if n == 0 {
			return succ, nil
		}

		succ += n
	}
}

// 同步
func (lgc *SinkLogic) sync2DB(num int32) (int, error) {
	defer logs.CatchPanic()

	keys, err := lgc.spopWatchKeys(num)
	if err != nil || len(keys) == 0 {
		return 0, err
	}

	logs.SInfo("%s keys len %d", lgc.name, len(keys))
	var list = make([]*CacheData, 0, len(keys))
	for _, v := range keys {
		data, err := lgc.getCacheFn(v)
		if err != nil {
			logs.SInfo("%s getData %s failed %s", lgc.name, v, err)
		}
		if data != nil {
			list = append(list, data)
		}
	}

	err = lgc.save2DBFn(list)
	return len(keys), err
}
