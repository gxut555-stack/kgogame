/**
 * @Author: Chen Bin
 * @File: frame.go
 * @Date: 2024-11-06 10:48:41
 * @Description:
 */

package ws

import (
	"errors"
	"fmt"
	"io"
	"kgogame/core/access/bigEndian"
	"net"
	"time"
)

/*  The following is websocket data frame:
+-+-+-+-+-------+-+-------------+-------------------------------+
0                   1                   2                   3   |
0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 |
+-+-+-+-+-------+-+-------------+-------------------------------+
|F|R|R|R| opcode|M| Payload len |    Extended payload length    |
|I|S|S|S|  (4)  |A|     (7)     |             (16/64)           |
|N|V|V|V|       |S|             |   (if payload len==126/127)   |
| |1|2|3|       |K|             |                               |
+-+-+-+-+-------+-+-------------+ - - - - - - - - - - - - - - - +
|     Extended payload length continued, if payload len == 127  |
+ - - - - - - - - - - - - - - - +-------------------------------+
|                               |Masking-key, if MASK set to 1  |
+-------------------------------+-------------------------------+
| Masking-key (continued)       |          Payload Data         |
+-------------------------------- - - - - - - - - - - - - - - - +
:                     Payload Data continued ...                :
+ - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - +
|                     Payload Data continued ...                |
+---------------------------------------------------------------+
WebSocket Opcode numbers are subject to the "Standards Action" IANA
  registration policy [RFC5226].

  IANA has added initial values to the registry as follows.

    |Opcode  | Meaning                             | Reference |
   -+--------+-------------------------------------+-----------|
    | 0      | Continuation Frame                  | RFC 6455  |
   -+--------+-------------------------------------+-----------|
    | 1      | Text Frame                          | RFC 6455  |
   -+--------+-------------------------------------+-----------|
    | 2      | Binary Frame                        | RFC 6455  |
   -+--------+-------------------------------------+-----------|
    | 8      | Connection Close Frame              | RFC 6455  |
   -+--------+-------------------------------------+-----------|
    | 9      | Ping Frame                          | RFC 6455  |
   -+--------+-------------------------------------+-----------|
    | 10     | Pong Frame                          | RFC 6455  |
   -+--------+-------------------------------------+-----------|
协议定义：
https://datatracker.ietf.org/doc/html/rfc6455#autoid-21
*/

// WsHeader represents websocket frame header.
// See https://tools.ietf.org/html/rfc6455#section-5.2
type WsHeader struct {
	Fin           bool
	Rsv1          bool
	Rsv2          bool
	Rsv3          bool
	OpCode        byte
	Masked        bool
	Mask          [4]byte
	HeaderLength  int32 //头长度
	PayloadLength int64 //长度
}

func (wh *WsHeader) Clear() {
	wh.Fin = false
	wh.Rsv1, wh.Rsv2, wh.Rsv3 = false, false, false
	wh.OpCode = 0
	wh.Masked = false
	wh.Mask = [4]byte{0, 0, 0, 0} //fast reset
	wh.HeaderLength = 0
	wh.PayloadLength = 0
}

func (wh *WsHeader) String() string {
	return fmt.Sprintf("fin:%t rsv1:%t rsv2:%t rsv3:%t opcode:%d mask:[0x%x,0x%x,0x%x,0x%x] masked:%t",
		wh.Fin, wh.Rsv1, wh.Rsv2, wh.Rsv3, wh.OpCode, wh.Mask[0], wh.Mask[1], wh.Mask[2], wh.Mask[3], wh.Masked)
}

// 解析包,payload必须大于2，此处不会做检测
func (wh *WsHeader) ParseHeader(r net.Conn, payload []byte) error {
	wh.Clear() //先清除
	wh.Fin = payload[0]&finalBit != 0
	wh.Rsv1 = payload[0]&rsv1Bit != 0
	wh.Rsv2 = payload[0]&rsv2Bit != 0
	wh.Rsv3 = payload[0]&rsv3Bit != 0
	wh.Masked = payload[1]&maskBit != 0
	wh.OpCode = payload[0] & 0x0F
	if wh.Rsv1 {
		return errors.New("rsv1 flag set , server not supported")
	}
	if wh.Rsv2 {
		return errors.New("rsv2 flag set , server not supported")
	}
	if wh.Rsv3 {
		return errors.New("rsv3 flag set , server not supported")
	}
	wh.HeaderLength += WEBSOCKET_HEADER_LEN
	wh.PayloadLength = int64(payload[1] & 0x7f)
	needRead := 0
	switch wh.PayloadLength {
	case 0x7e: //126,需要继续读2个字节
		needRead = 2
	case 0x7f: //127，需要继续8个字节
		needRead = 8
	default:
		if wh.PayloadLength > 0x7f {
			return errors.New("header error: unexpected payload length bits")
		}
	}
	if wh.Masked {
		needRead += 4
	}
	if needRead > 0 {
		pos := int(wh.HeaderLength)
		if _, err := io.ReadAtLeast(r, payload[pos:pos+needRead], needRead); err != nil {
			return fmt.Errorf("read header extend length error: %s", err)
		}
		//解析长度
		if wh.PayloadLength == 0x7e {
			wh.PayloadLength = int64(bigEndian.ReadUint16(payload[pos : pos+2]))
			pos += 2
		} else if wh.PayloadLength == 0x7f {
			wh.PayloadLength = int64(bigEndian.ReadUint64(payload[pos : pos+8]))
			pos += 8
		}
		if wh.Masked {
			copy(wh.Mask[:], payload[pos:pos+4])
		}
	}
	wh.HeaderLength += int32(needRead)
	return nil
}

const (
	WEBSOCKET_OPCODE_CONTINUATION_FRAME = 0x0
	WEBSOCKET_OPCODE_TEXT_FRAME         = 0x1
	WEBSOCKET_OPCODE_BINARY_FRAME       = 0x2
	WEBSOCKET_OPCODE_CONNECTION_CLOSE   = 0x8
	WEBSOCKET_OPCODE_PING               = 0x9
	WEBSOCKET_OPCODE_PONG               = 0xa
)

const (
	// Frame header byte 0 bits from Section 5.2 of RFC 6455
	finalBit = 1 << 7
	rsv1Bit  = 1 << 6
	rsv2Bit  = 1 << 5
	rsv3Bit  = 1 << 4

	// Frame header byte 1 bits from Section 5.2 of RFC 6455
	maskBit = 1 << 7

	WEBSOCKET_HEADER_LEN       = 2         //ws 头长度
	maxFrameHeaderSize         = 2 + 8 + 4 // Fixed header + length + mask
	maxControlFramePayloadSize = 125

	writeWait = time.Second

	defaultReadBufferSize  = 4096
	defaultWriteBufferSize = 4096

	continuationFrame = 0
	noFrame           = -1
)

// Close codes defined in RFC 6455, section 11.7.
const (
	CloseNormalClosure           = 1000
	CloseGoingAway               = 1001
	CloseProtocolError           = 1002
	CloseUnsupportedData         = 1003
	CloseNoStatusReceived        = 1005
	CloseAbnormalClosure         = 1006
	CloseInvalidFramePayloadData = 1007
	ClosePolicyViolation         = 1008
	CloseMessageTooBig           = 1009
	CloseMandatoryExtension      = 1010
	CloseInternalServerErr       = 1011
	CloseServiceRestart          = 1012
	CloseTryAgainLater           = 1013
	CloseTLSHandshake            = 1015
)

// ErrCloseSent is returned when the application writes a message to the
// connection after sending a close message.
var ErrCloseSent = errors.New("websocket: close sent")

// ErrReadLimit is returned when reading a message that is larger than the
// read limit set for the connection.
var ErrReadLimit = errors.New("websocket: read limit exceeded")

// for test
func NewWsFrameText() *WsHeader {
	return &WsHeader{
		Fin:    true,
		OpCode: WEBSOCKET_OPCODE_TEXT_FRAME,
		Masked: false,
	}
}

func NewWsFrameBinary() *WsHeader {
	return &WsHeader{
		Fin:    true,
		OpCode: WEBSOCKET_OPCODE_BINARY_FRAME,
		Masked: false,
	}
}

func EncodeWsHeaderText(payload []byte) ([]byte, int) {
	header := NewWsFrameText()
	header.PayloadLength = int64(len(payload))
	buff := make([]byte, maxFrameHeaderSize)
	header.HeaderLength = WEBSOCKET_HEADER_LEN
	if header.Fin {
		buff[0] |= finalBit
	}
	buff[0] |= header.OpCode & 0x0F
	if header.PayloadLength < 126 {
		buff[1] = byte(header.PayloadLength)
	} else if header.PayloadLength < 65536 {
		buff[1] = 126
		bigEndian.WriteUint16(buff[2:4], uint16(header.PayloadLength))
		header.HeaderLength += 2
	} else {
		buff[1] = 127
		bigEndian.WriteUint64(buff[2:10], uint64(header.PayloadLength))
		header.HeaderLength += 8
	}
	return buff[0:header.HeaderLength], int(header.HeaderLength)
}

func EncodeWsHeaderBinary(payload []byte) ([]byte, int) {
	header := NewWsFrameBinary()
	header.PayloadLength = int64(len(payload))
	buff := make([]byte, maxFrameHeaderSize)
	header.HeaderLength = WEBSOCKET_HEADER_LEN
	if header.Fin {
		buff[0] |= finalBit
	}
	buff[0] |= header.OpCode & 0x0F
	if header.PayloadLength < 126 {
		buff[1] = byte(header.PayloadLength)
	} else if header.PayloadLength < 65536 {
		buff[1] = 126
		bigEndian.WriteUint16(buff[2:4], uint16(header.PayloadLength))
		header.HeaderLength += 2
	} else {
		buff[1] = 127
		bigEndian.WriteUint64(buff[2:10], uint64(header.PayloadLength))
		header.HeaderLength += 8
	}
	return buff[0:header.HeaderLength], int(header.HeaderLength)
}
