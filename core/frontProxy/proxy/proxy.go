package proxy

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"kgogame/util/plog"
	"net"
	"strconv"
	"strings"
)

// Proxy - Manages a Proxy connection, piping data between local and remote.
type Proxy struct {
	sentBytes     uint64
	receivedBytes uint64
	laddr, raddr  *net.TCPAddr
	lconn, rconn  io.ReadWriteCloser
	erred         bool
	errsig        chan bool
	tlsUnwrapp    bool
	tlsAddress    string

	Matcher  func([]byte)
	Replacer func([]byte) []byte

	// Settings
	Nagles bool
	//Log       Logger
	OutputHex bool

	//Client Ip
	clientIp   string
	clientPort uint16
}

// New - Create a new Proxy instance. Takes over local connection passed in,
// and closes it when finished.
func New(lconn *net.TCPConn, laddr, raddr *net.TCPAddr) *Proxy {
	return &Proxy{
		lconn:  lconn,
		laddr:  laddr,
		raddr:  raddr,
		erred:  false,
		errsig: make(chan bool),
		//Log:    NullLogger{},
	}
}

// NewTLSUnwrapped - Create a new Proxy instance with a remote TLS server for
// which we want to unwrap the TLS to be able to connect without encryption
// locally
func NewTLSUnwrapped(lconn *net.TCPConn, laddr, raddr *net.TCPAddr, addr string) *Proxy {
	p := New(lconn, laddr, raddr)
	p.tlsUnwrapp = true
	p.tlsAddress = addr
	return p
}

type setNoDelayer interface {
	SetNoDelay(bool) error
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

// Get the Client ip port
func (p *Proxy) getClientIp() error {

	if p.lconn == nil {
		return errors.New("net.TCPConn == nil")
	}

	//client ip addr
	addr := p.lconn.(*net.TCPConn).RemoteAddr().String()
	if pos := strings.LastIndex(addr, ":"); pos > -1 {
		p.clientIp = addr[:pos]
		if pos < len(addr)-1 {
			if port, err := strconv.Atoi(addr[pos+1:]); err == nil {
				p.clientPort = uint16(port)
			}
		}
	} else {
		p.clientIp = addr
	}

	return nil
}

func IPToUInt32(ipnr net.IP) uint32 {
	bits := strings.Split(ipnr.String(), ".")

	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])

	var sum uint32

	sum += uint32(b0) << 24
	sum += uint32(b1) << 16
	sum += uint32(b2) << 8
	sum += uint32(b3)

	return sum
}

func UInt32ToIP(intIP uint32) net.IP {
	var bytes [4]byte
	bytes[0] = byte(intIP & 0xFF)
	bytes[1] = byte((intIP >> 8) & 0xFF)
	bytes[2] = byte((intIP >> 16) & 0xFF)
	bytes[3] = byte((intIP >> 24) & 0xFF)

	return net.IPv4(bytes[3], bytes[2], bytes[1], bytes[0])
}

// Start - open connection to remote and start proxying data.
func (p *Proxy) Start() {
	defer p.lconn.Close()

	var err error
	//connect to remote
	if p.tlsUnwrapp {
		p.rconn, err = tls.Dial("tcp", p.tlsAddress, nil)
	} else {
		p.rconn, err = net.DialTCP("tcp", nil, p.raddr)
	}
	if err != nil {
		plog.VWarn("Remote connection failed: %s", err)
		return
	}
	defer p.rconn.Close()

	//nagles?
	if p.Nagles {
		if conn, ok := p.lconn.(setNoDelayer); ok {
			conn.SetNoDelay(true)
		}
		if conn, ok := p.rconn.(setNoDelayer); ok {
			conn.SetNoDelay(true)
		}
	}

	err = p.getClientIp()
	if err != nil {
		plog.VWarn("getClientIp failed: %s", err)
		return
	}

	//display both ends
	plog.VInfo("Opened %s [%s:%d -> %s]", p.laddr.String(), p.clientIp, p.clientPort, p.raddr.String())

	ClientIP, Type := ParseIP(p.clientIp)

	//plog.VInfo("IP:[%s][%d][%d]", ClientIP, Type, len(ClientIP))

	//send proxy protocol
	pp := &ProxyProtocol{}
	pp.Sig = V2sig
	pp.Vercmd = 0x20 | 0x01
	if Type == 4 {
		pp.Fam = 0x11
		pp.len = 12
		pp.V4 = &V4Protocal{}
		pp.V4.Src_addr = IPToUInt32(ClientIP)
		pp.V4.Src_port = p.clientPort

		//plog.VInfo("Src_addr [%d][%d]", pp.V4.Src_addr, pp.V4.Src_port)

	} else if Type == 6 {
		pp.Fam = 0x21
		pp.len = 36
		pp.V6 = &V6Protocal{}

		for k, v := range ClientIP.String() {

			//For protect!!
			if k > 15 {
				break
			}
			pp.V6.Src_addr[k] = byte(v)
		}
		pp.V6.Src_port = p.clientPort
	}

	err, SendBuf := EncodeProxyProtocol(pp)

	if err != nil {
		plog.VWarn("EncodeProxyProtocol failed: %s", err)
	} else {
		//Send to real Server
		n, err := p.rconn.Write(SendBuf)
		if err != nil {
			plog.VWarn("Write failed %s, %d", err, n)
			return
		}
	}

	//bidirectional copy
	go p.pipe(p.lconn, p.rconn, fmt.Sprintf("[%s:%d -> %s]", ClientIP, p.clientPort, p.raddr))
	go p.pipe(p.rconn, p.lconn, fmt.Sprintf("[%s -> %s:%d]", p.raddr, ClientIP, p.clientPort))

	//wait for close...
	<-p.errsig
	plog.VInfo("Closed [%s:%d -> %s] (%d bytes sent, %d bytes recieved)", ClientIP, p.clientPort, p.raddr, p.sentBytes, p.receivedBytes)
}

func (p *Proxy) err(s string, err error) {
	if p.erred {
		return
	}
	//if err != io.EOF {
	plog.VWarn(s, err)
	//}
	p.errsig <- true
	p.erred = true
}

func (p *Proxy) pipe(src, dst io.ReadWriter, prefix string) {
	islocal := src == p.lconn

	var dataDirection string
	if islocal {
		dataDirection = ">>> %d bytes sent%s"
	} else {
		dataDirection = "<<< %d bytes recieved%s"
	}

	var byteFormat string
	if p.OutputHex {
		byteFormat = "%x"
	} else {
		byteFormat = "%s"
	}

	//directional copy (64k buffer)
	buff := make([]byte, 0xffff)
	for {
		n, err := src.Read(buff)
		if err != nil {
			p.err(prefix+"Read failed '%s'\n", err)
			return
		}
		b := buff[:n]

		//execute match
		if p.Matcher != nil {
			p.Matcher(b)
		}

		//execute replace
		if p.Replacer != nil {
			b = p.Replacer(b)
		}

		//show output
		plog.VDebug(dataDirection, n, "")
		plog.VInfo(byteFormat, b)

		//write out result
		n, err = dst.Write(b)
		if err != nil {
			p.err(prefix+"Write failed '%s'\n", err)
			return
		}
		if islocal {
			p.sentBytes += uint64(n)
		} else {
			p.receivedBytes += uint64(n)
		}
	}
}
