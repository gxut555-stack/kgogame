package plog

import (
	"errors"
	"fmt"
	"net"
)

const MSG_MAX_SIZE = 380730

type UdpLogger struct {
	addr    *net.UDPAddr
	conn    *net.UDPConn
	topic   string
	maxSize int
}

func CreateLoggerFromAddr(addr string, topic string) (*UdpLogger, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("net.ResolveUDPAddr(udp, %s) error: %s", addr, err.Error())
	}
	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, fmt.Errorf("net.DialUDP(udp, nil, %s) error: %s", addr, err.Error())
	}
	if err = udpConn.SetWriteBuffer(MSG_MAX_SIZE); err != nil {
		return nil, fmt.Errorf("udpConn.SetWriteBuffer(%d) error: %s", MSG_MAX_SIZE, err.Error())
	}
	logger := &UdpLogger{
		addr:    udpAddr,
		conn:    udpConn,
		topic:   topic,
		maxSize: MSG_MAX_SIZE - len(topic) - 1,
	}
	return logger, nil
}

func (l *UdpLogger) Write(p []byte) (int, error) {
	if len(p) > l.maxSize {
		fmt.Println(l.topic + "|" + string(p))
		return 0, errors.New("message is too large")
	}
	return l.conn.Write([]byte(l.topic + "|" + string(p)))
}
