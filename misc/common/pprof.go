package common

import (
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
)

// 浏览器会限制以下端口访问
// 1719,   // h323gatestat
// 1720,   // h323hostcall
// 1723,   // pptp
// 2049,   // nfs
// 3659,   // apple-sasl / PasswordServer
// 4045,   // lockd
// 5060,   // sip
// 5061,   // sips
// 6000,   // X11
// 6566,   // sane-port
// 6665,   // Alternate IRC [Apple addition]
// 6666,   // Alternate IRC [Apple addition]
// 6667,   // Standard IRC [Apple addition]
// 6668,   // Alternate IRC [Apple addition]
// 6669,   // Alternate IRC [Apple addition]
// 6697,   // IRC + TLS
// 10080,  // Amanda
func EnablePProf(ports ...int) {
	var port = 0
	if len(ports) > 0 {
		port = ports[0]
	}
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Printf("start pprof fail %s", err)
		return
	}

	fmt.Println("enable pprof port:", listener.Addr().(*net.TCPAddr).Port)
	if err = http.Serve(listener, nil); err != nil {
		log.Printf("pprof fail %s", err)
	}
}
