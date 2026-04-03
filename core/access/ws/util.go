/**
 * @Author: Chen Bin
 * @File: util.go
 * @Date: 2024-11-06 10:49:13
 * @Description:
 */

package ws

import (
	"net"
	"net/http"
	"strings"
)

// https://developers.cloudflare.com/fundamentals/reference/http-headers/#x-forwarded-for
// cloudflare 推荐使用此值,使用http 头读值无论是用xxf、还是xri都会存在被伪造的情况,包括ip层也是可以加层转发而导致拿不到真实ip
const (
	CCI = "Cf-Connecting-Ip"
	XXF = "X-Forwarded-For"
	XRI = "X-Real-IP"
)

func GetRemoteIp(req *http.Request) string {
	//优先读取cci
	if val := req.Header.Get(CCI); val != "" {
		if net.ParseIP(val) != nil { //有效ip
			return val
		}
	}
	//再次读取xxf
	if val := req.Header.Get(XXF); val != "" {
		//由于golang标准库会把X-FORWARDED-FOR: 192.168.10.1, 192.168.10.1当作一条处理,故需要判断,
		//同时行业约定所有proxy转发请求时建议append到X-FORWARDED-FOR后面，故第一个ip即原始ip
		if pos := strings.IndexByte(val, ','); pos != -1 {
			realIp := val[:pos]
			if net.ParseIP(realIp) != nil {
				return realIp
			}
		} else {
			if net.ParseIP(val) != nil { //有效ip
				return val
			}
		}
	}
	//再次读取X-Real-IP
	if val := req.Header.Get(XRI); val != "" {
		if net.ParseIP(val) != nil { //有效ip
			return val
		}
	}
	return ""
}
func ParseIP(s string) (net.IP, int) {

	//If the string contain semicolon?
	if lindex := strings.Index(s, "["); lindex > -1 {
		s = s[lindex:]
	}
	if rindex := strings.Index(s, "]"); rindex > -1 {
		s = s[:rindex]
	}

	ip := net.ParseIP(s)
	if ip == nil {
		return nil, 0
	}

	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '.':
			return ip, 4
		case ':':
			return ip, 6
		}
	}
	return nil, 0
}
