// 配置中心站，从命令行启动参数或者配置文件或者网络服务中获取配置信息，先放在此处，然后向各个模块分发
// eg: plog.InitConf(xxx,yyy,zzz) 分发到plog的配置中去
package gconf

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"kgogame/util/gconf/gconfclient"
	"kgogame/util/gconf/gconftypes"
	"net/url"
	"strings"
	"sync"
)

type ApolloConf struct {
	Name    string
	Type    string
	Content string
}

type CustomAgolloConfig struct {
	Custom map[string]string `json:"custom"`
}

// 版本和svrid信息
type CmdConfig struct {
	IpAddress     string //rpc服务监听的ip端口
	Version       string //rpc版本
	ServerID      int32  //rpc的svrid
	ServerName    string //rpc服务名称
	Level         int32  //gamesvr使用的场次信息（仅供gamesvr使用，其他服可无视）
	Type          string //标识此game的类型 “M”标识比赛用game “H”标识大厅用game
	AcPort        int    //ac服务对外放开的端口
	AgolloCfgFile string //Agollo的配置文件路径
	PLogConf      string //plog的ini配置文件路径
	TomlConf      string //toml配置文件
	GameId        int32  //游戏id
}

// redis连接池的配置
type RedisConf struct {
	Host           string //IP地址:端口号
	Pass           string //密码
	MaxIdle        int    //最大空闲连接数
	MaxActive      int    //最大连接数(包含了空闲的)
	IdleTimeout    int    //池中的连接空闲多久之后超时（单位：秒）
	ConnectTimeout int    //连接超时时间（单位:毫秒）
	ReadTimeout    int    //读超时时间（单位:毫秒）
	WriteTimeout   int    //写超时时间（单位:毫秒）
	Database       int    //数据库编号
}

type DBConf struct {
	DBCount    int32    `json:"DBCount" toml:"DBCount"`
	DSN        []string `json:"DSN" toml:"DSN"`
	TableCount int32    `json:"TableCount" toml:"TableCount"`
	Charset    string   `json:"Charset" toml:"Charset"`
	Database   string   `json:"Database" toml:"Database"`
}

/*
	MongoDB配置信息
	需要在连接串中对特殊字符进行转义处理，转义规则如下：

! : %21
@ : %40
# : %23
$ : %24
% : %25
^ : %5e
& : %26
* : %2a
( : %28
) : %29
_ : %5f
+ : %2b
= : %3d
*/
type MongoConfig struct {
	Name           string   `json:"Name"`
	DBCount        int      `json:"DBCount"` //DB连接数
	DSN            []string `json:"DSN"`
	Database       string   `json:"Database"`
	ConnectTimeout int      `json:"ConnectTimeout"` //连接超时
}

var (
	CmdConf        *CmdConfig
	apolloConfLock sync.RWMutex

	redisPoolConf map[string]*RedisConf   //redis全局连接池配置
	mysqlConf     map[string]*DBConf      //mysql 库表对应配置
	mongoConf     map[string]*MongoConfig //mongodb 库表对应配置

	isApolloInit bool //阿波罗只初始化一次
	initMetux    sync.Mutex

	gConfAddr    string
	gConfProject string
	gConfCli     *gconfclient.GConfClient
)

func init() {
	CmdConf = new(CmdConfig)
	redisPoolConf = make(map[string]*RedisConf) //redis配置
	mysqlConf = make(map[string]*DBConf)        //数据库配置
	mongoConf = make(map[string]*MongoConfig)   //MongoDB配置
}

func InitApolloConf(apolloPath string) error {
	var (
		addr, project string
	)
	//判断是不是新的gconf
	if len(apolloPath) < 6 || apolloPath[0:6] != "tcp://" {
		return fmt.Errorf("unsuppoted agollo value: %s", apolloPath)
	}

	if data, err := url.Parse(apolloPath); err != nil {
		return fmt.Errorf("invalid agollo value: %s, parse error: %s", apolloPath, err.Error())
	} else if data.Host == "" || data.Path == "" {
		return fmt.Errorf("invalid agollo value: %s", apolloPath)
	} else {
		addr, project = data.Host, strings.TrimLeft(data.Path, "/")
	}

	initMetux.Lock()
	if isApolloInit {
		initMetux.Unlock()
		return nil
	}
	gConfAddr, gConfProject = addr, project
	gConfCli = gconfclient.NewGConfClient()
	if err := gConfCli.Start(gConfAddr, gConfProject, 0, gConfCallback); err != nil {
		initMetux.Unlock()
		return fmt.Errorf("gConfCli.Start(%s, %s, 0, gConfCallback) error: %s", gConfAddr, gConfProject, err.Error())
	}
	isApolloInit = true
	initMetux.Unlock()

	if err := gConfRefresh(gConfProject); err != nil {
		return fmt.Errorf("gConfRefresh(%s) error: %s", gConfProject, err.Error())
	}
	return nil
}

func InitConf() error {
	//加载命令行启动参数
	LoadConf()
	err := ValidConf()
	if err != nil {
		fmt.Println("加载启动参数失败，err:", err)
		return err
	}

	//加载apollo配置
	err = InitApolloConf(CmdConf.AgolloCfgFile)
	if err != nil {
		fmt.Println("初始化阿波罗配置内容失败,err:", err)
		return err
	}
	return err
}

// 加载启动参数
func LoadConf() {
	AcPort := flag.Int("acPort", 0, "access Listen Port")
	Ver := flag.String("v", "0.0.0.0", "Version")
	Name := flag.String("name", "", "Name")
	ServID := flag.Int("servId", 0, "servId")
	IpAddress := flag.String("h", "", "IpAddress")
	Level := flag.Int("l", 0, "Level")
	Type := flag.String("t", "H", "game类型 M-比赛 H-大厅") //默认为大厅使用的game
	agolloCfgFile := flag.String("agollo", "", "Config file of apollo")
	plogConf := flag.String("plogCfg", "", "plog config file(ini)")
	tomlConf := flag.String("toml", "", "toml config file")
	gameId := flag.Int("gameId", 0, "game id")

	flag.Parse()

	//赋值
	CmdConf.Version = *Ver
	CmdConf.ServerID = int32(*ServID)
	CmdConf.ServerName = *Name
	CmdConf.Level = int32(*Level)
	CmdConf.IpAddress = *IpAddress
	CmdConf.Type = *Type
	CmdConf.AcPort = *AcPort
	CmdConf.AgolloCfgFile = *agolloCfgFile
	CmdConf.PLogConf = *plogConf
	CmdConf.TomlConf = *tomlConf
	CmdConf.GameId = int32(*gameId)

}

func ValidConf() error {
	//合法性校验
	if len(strings.Split(CmdConf.Version, ".")) != 4 {
		return errors.New("missing argument '-v' (version)")
	}
	if strings.Index(CmdConf.IpAddress, ":") < 0 {
		return errors.New("missing argument '-h' (listen)")
	}
	if CmdConf.ServerName == "" {
		return errors.New("missing argument '-name' (server name)")
	}
	if CmdConf.ServerID == 0 {
		return errors.New("missing argument '-servId' (server id)")
	}
	return nil
}

// 重新加载指定配置项
func gConfRefresh(project string) error {
	if items, ok := gConfCli.ProjectItems(project); ok {
		for k, item := range items {
			switch item.Format {
			case "redis":
				ob := new(RedisConf)
				if err := json.Unmarshal([]byte(item.Value), ob); err != nil {
					return fmt.Errorf("item '%s' is not a invlid redis config: %s", k, err.Error())
				} else {
					apolloConfLock.Lock()
					redisPoolConf[k] = ob
					apolloConfLock.Unlock()
				}
			case "mysql":
				ob := new(DBConf)
				if err := json.Unmarshal([]byte(item.Value), ob); err != nil {
					return fmt.Errorf("item '%s' is not a invalid db config: %s", k, err)
				} else {
					apolloConfLock.Lock()
					mysqlConf[k] = ob
					apolloConfLock.Unlock()
				}
			}
		}
	}
	return nil
}

func gConfCallback(project string, version int32, old map[string]gconftypes.Item) {
	if err := gConfRefresh(project); err != nil {
		fmt.Printf("gConfRefresh(%s) error: %s", project, err.Error())
	}
}
