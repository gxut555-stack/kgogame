package getConfig

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
)

func GetLocalIP() string {
	var localIP string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				localIP = ipnet.IP.String()
			}
		}
	}
	return localIP
}

/*
func GetConfFile(ServiceName string) ([]byte, error) {

		req := new(protoConfig.OperationReq)
		req.ServiceName = ServiceName

		reply := &protoConfig.OperationRsp{}
		ctx, _ := context.WithTimeout(context.Background(), 3*time.Second) //超时控制
		err := rpcxClient.XCall(ctx, "ConfigServer", "Get", req, reply)
		if err != nil {
			//plog.VFatal("[err happened]: %v", err)
			return nil, err
		} else {
			if reply.Code < 0 {
				return nil, errors.New(string(reply.Files))
			}
		}

		return reply.Files, nil
	}
*/
func SetFile(ServiceName string, fileStr []byte) error {
	err := ioutil.WriteFile(ServiceName, fileStr, 0644)

	if err != nil {
		return err
	}
	return nil
}

/*
func GetSetConfigureFileAndParse(confFileName string) (*IniParser, error) {

	iniPs := new(IniParser)
	configFile, errConf := GetConfFile(confFileName)
	if errConf != nil {
		return iniPs, errConf
	}

	errConf = SetFile(confFileName, configFile)
	if errConf != nil {
		return iniPs, errConf
	}

	errConf = iniPs.Load(confFileName)
	if errConf != nil {
		return iniPs, errConf
	}

	return iniPs, nil
}
*/

func GetLocalIpAndPort() (string, int, error) {

	sIP := GetLocalIP()
	if sIP == "" {
		return "", 0, errors.New("GetLocalIP error")
	}

	port := flag.Int("port", 0, "listen port")
	flag.Parse()

	if *port == 0 {
		return "", 0, errors.New("usage: ./accessServer -port=...")
	}

	return sIP, *port, nil
}
