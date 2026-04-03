/**
 * @Author:
 * @File: protocol.go
 * @Date: 2024-12-24 17:16:15
 * @Description: https://www.haproxy.org/download/1.8/doc/proxy-protocol.txt
 */

package proxyProtocol

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"kgogame/core/access/bigEndian"
	"kgogame/util/logs"
	"net"
	"net/netip"
)

var (
	V2sig = [12]byte{'\x0D', '\x0A', '\x0D', '\x0A', '\x00', '\x0D', '\x0A', '\x51', '\x55', '\x49', '\x54', '\x0A'}
)

type V4Protocol struct {
	SrcAddr uint32
	DstAddr uint32
	SrcPort uint16
	DstPort uint16
}

// Ipv6
type V6Protocol struct {
	SrcAddr [16]byte
	DstAddr [16]byte
	SrcPort uint16
	DstPort uint16
}

type Tlv struct {
}

type ProxyProtocol struct {
	Sig    [12]byte
	VerCmd byte
	Fam    byte
	len    uint16
	V4     *V4Protocol
	V6     *V6Protocol
	Tlv    *Tlv //标准里面还支持tlv,但暂不支持
}

// 获取ip地址
func (p *ProxyProtocol) GetIpPort() (string, int32, string) {
	var (
		ip   string
		port int32
		addr string
	)
	if p.Fam == 0x11 {
		ip = UInt32ToIP(p.V4.SrcAddr).String()
		port = int32(p.V4.SrcPort)
		addr = fmt.Sprintf("%s:%d", ip, port)
	} else if p.Fam == 0x21 {
		ip = netip.AddrFrom16(p.V6.SrcAddr).String()
		logs.Info("ipv6 bytes :%#v", p.V6.SrcAddr)
		port = int32(p.V6.SrcPort)
		addr = fmt.Sprintf("[%s]:%d", ip, port)
	}
	return ip, port, addr
}

// 协议关键头是否相等
func Cmp(buffer []byte) bool {
	if len(buffer) < 12 {
		return false
	}
	if bytes.Compare(buffer, V2sig[:]) == 0 {
		return true
	}
	return false
}

/**
c source struct
struct {
	  uint8_t sig[12];
	  uint8_t ver_cmd;
	  uint8_t fam;
	  uint16_t len;
	  union {
		struct {  // for TCP/UDP over IPv4, len = 12
			uint32_t src_addr;
			uint32_t dst_addr;
			uint16_t src_port;
			uint16_t dst_port;
		} ip4;
		struct {  // for TCP/UDP over IPv6, len = 36
			uint8_t  src_addr[16];
			uint8_t  dst_addr[16];
			uint16_t src_port;
			uint16_t dst_port;
		} ip6;
		struct {  // for AF_UNIX sockets, len = 216
			uint8_t src_addr[108];
			uint8_t dst_addr[108];
		} unx;
	} addr;
} v2;
**/

// 暂时不实现
func EncodeProxyProtocol(pp *ProxyProtocol) (error, []byte) {
	return errors.New("not implement"), nil
}

// Decode the protocol
func DecodeProxyProtocol(buff []byte, bio *bufio.Reader) (*ProxyProtocol, int, error) {
	if len(buff) != 16 {
		return nil, 0, errors.New("buff too short")
	}
	pos := 12
	pp := &ProxyProtocol{}
	pp.VerCmd = buff[pos]
	pp.Fam = buff[pos+1]
	pp.len = bigEndian.ReadUint16(buff[pos+2 : pos+4])

	if pp.VerCmd&0xF0 != 0x20 {
		return nil, 0, errors.New("invalid version cmd")
	}
	proxyPacketLength := 16 + int(pp.len)
	moreBuffer, err := bio.Peek(proxyPacketLength)
	if err != nil {
		return nil, 0, err
	}
	isHasTlv := false
	n := 16
	switch pp.Fam {
	case 0x11:
		{ //IpV4
			pp.V4 = &V4Protocol{}
			if len(moreBuffer) < 28 {
				return nil, 0, fmt.Errorf("error ipV4 protocol moreBuffer:%#v pp:%#v", moreBuffer, pp)
			}
			pp.V4.SrcAddr = bigEndian.ReadUint32(moreBuffer[n:])
			n += 4
			pp.V4.DstAddr = bigEndian.ReadUint32(moreBuffer[n:])
			n += 4
			pp.V4.SrcPort = bigEndian.ReadUint16(moreBuffer[n:])
			n += 2
			pp.V4.DstPort = bigEndian.ReadUint16(moreBuffer[n:])
			n += 2
			if n < len(moreBuffer) {
				isHasTlv = true
			}
		}
	case 0x21:
		{ //IpV6
			pp.V6 = &V6Protocol{}
			if len(moreBuffer) < 52 {
				return nil, 0, errors.New("error ipV6 protocol")
			}
			copy(pp.V6.SrcAddr[:], moreBuffer[n:n+16])
			n += 16
			copy(pp.V6.DstAddr[:], moreBuffer[n:n+16])
			n += 16
			pp.V6.SrcPort = bigEndian.ReadUint16(moreBuffer[n:])
			n += 2
			pp.V6.DstPort = bigEndian.ReadUint16(moreBuffer[n:])
			n += 2
			if n < len(moreBuffer) {
				isHasTlv = true
			}
		}
	default:
		return nil, 0, errors.New("not ipv4 or ipv6")
	}
	if isHasTlv {
		//暂时不支持解析tlv
	}
	return pp, proxyPacketLength, nil
}

func UInt32ToIP(intIP uint32) net.IP {
	var b [4]byte
	b[0] = byte(intIP & 0xFF)
	b[1] = byte((intIP >> 8) & 0xFF)
	b[2] = byte((intIP >> 16) & 0xFF)
	b[3] = byte((intIP >> 24) & 0xFF)

	return net.IPv4(b[3], b[2], b[1], b[0])
}
