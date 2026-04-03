package proxy

import (
	"encoding/binary"
	"errors"
)

/*
	This protocol is based on https://www.haproxy.org/download/1.8/doc/proxy-protocol.txt
*/

var (
	V2sig = [12]byte{'\x0D', '\x0A', '\x0D', '\x0A', '\x00', '\x0D', '\x0A', '\x51', '\x55', '\x49', '\x54', '\x0A'}
)


//Ipv4
type V4Protocal struct {
	Src_addr	uint32
	Dst_addr	uint32
	Src_port	uint16
	Dst_port	uint16
}


//Ipv6
type V6Protocal struct {
	Src_addr	[16]byte
	Dst_addr	[16]byte
	Src_port	uint16
	Dst_port	uint16
}


type ProxyProtocol struct {
	Sig 	[12]byte
	Vercmd	byte
	Fam 	byte
	len		uint16
	V4		*V4Protocal
	V6		*V6Protocal
}


func EncodeProxyProtocol(pp *ProxyProtocol) (error, []byte) {
	if pp == nil {
		return errors.New("Nil object"), nil
	}

	var length int
	var buff []byte

	//打包
	if pp.V4 != nil {
		length = 12 + 1 + 1 + 2 + 12
		buff = make([]byte, length)

		//copy(buff[:], *(*[]byte)(unsafe.Pointer(&pp.Sig)))
		//n = n + 12

		var n int
		for _,v :=range pp.Sig {
			buff[n] = v
			n = n + 1
		}

		buff[n] = pp.Vercmd
		n = n + 1

		buff[n] = pp.Fam
		n = n + 1

		binary.BigEndian.PutUint16(buff[n:], pp.len)
		n = n + 2

		binary.BigEndian.PutUint32(buff[n:], pp.V4.Src_addr)
		n = n + 4

		binary.BigEndian.PutUint32(buff[n:], pp.V4.Dst_addr)
		n = n + 4

		binary.BigEndian.PutUint16(buff[n:], pp.V4.Src_port)
		n = n + 2

		binary.BigEndian.PutUint16(buff[n:], pp.V4.Dst_port)
		n = n + 2


	} else if pp.V6 != nil {
		length = 12 + 1 + 1 + 2 + 36

		buff = make([]byte, length)

		//copy(buff[:], *(*[]byte)(unsafe.Pointer(&pp.Sig)))
		//n = n + 12

		var n int
		for _,v :=range pp.Sig {
			buff[n] = v
			n = n + 1
		}

		buff[n] = pp.Vercmd
		n = n + 1

		buff[n] = pp.Fam
		n = n + 1

		binary.BigEndian.PutUint16(buff[n:], pp.len)
		n = n + 2

		for _,v := range pp.V6.Src_addr {
			buff[n] = v
			n = n + 1
		}
		//binary.BigEndian.PutUint32(buff[n:], pp.v6.src_addr)
		//n = n + 4

		for _,v := range pp.V6.Dst_addr {
			buff[n] = v
			n = n + 1
		}
		//binary.BigEndian.PutUint32(buff[n:], pp.v6.dst_addr)
		//n = n + 4

		binary.BigEndian.PutUint16(buff[n:], pp.V6.Src_port)
		n = n + 2

		binary.BigEndian.PutUint16(buff[n:], pp.V6.Dst_port)
		n = n + 2

	} else {
		return errors.New("bad v4,v6"), nil
	}

	return nil, buff
}


//Decode the protocol
func DecodeProxyProtocol(buff []byte) (*ProxyProtocol, error){

	if buff == nil {
		return nil, errors.New("nil buff")
	}

	if len(buff) < 16 {
		return nil, errors.New("buff too short")
	}

	var n int

	n = n + 12
	pp := &ProxyProtocol{}

	for k,v := range buff[:n] {
		pp.Sig[k] = v
	}

	pp.Vercmd = buff[n]
	n = n + 1

	pp.Fam = buff[n]
	n = n + 1

	pp.len = binary.BigEndian.Uint16(buff[n:])
	n = n + 2

	//ipV4
	if pp.Fam == 0x11 {
		if pp.len == 12 {

			if len(buff) < 28 {
				return nil, errors.New("error ipV4 protocol buff too short")
			}

			ipV4 := &V4Protocal{}
			pp.V4 = ipV4
			pp.V4.Src_addr = binary.BigEndian.Uint32(buff[n:])
			n = n + 4
			pp.V4.Dst_addr = binary.BigEndian.Uint32(buff[n:])
			n = n + 4
			pp.V4.Src_port = binary.BigEndian.Uint16(buff[n:])
			n = n + 2
			pp.V4.Dst_port = binary.BigEndian.Uint16(buff[n:])
			n = n + 2

		} else {
			return nil, errors.New("error ipV4 protocol")
		}
	} else if pp.Fam == 0x21 {
		if pp.len == 36 {

			if len(buff) < 52 {
				return nil, errors.New("error ipV6 protocol buff too short")
			}

			ipV6 := &V6Protocal{}
			pp.V6 = ipV6

			for k,_ := range pp.V6.Src_addr {
				pp.V6.Src_addr[k] = buff[n + k]
			}
			n = n + 16

			for k,_ := range pp.V6.Dst_addr {
				pp.V6.Dst_addr[k] = buff[n + k]
			}
			n = n + 16

			pp.V6.Src_port = binary.BigEndian.Uint16(buff[n:])
			n = n + 2
			pp.V6.Dst_port = binary.BigEndian.Uint16(buff[n:])
			n = n + 2

		} else {
			return nil, errors.New("error ipV6 protocol")
		}
	} else {
		return nil, errors.New("error v4 or v6")
	}


	return pp, nil
}