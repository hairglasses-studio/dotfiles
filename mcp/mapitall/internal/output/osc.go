package output

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"sync"

	"github.com/hairglasses-studio/mapping"
)

// OSCTarget sends OSC messages over UDP.
type OSCTarget struct {
	mu    sync.Mutex
	conns map[string]*net.UDPConn // host:port -> conn
}

// NewOSCTarget creates an OSC output target.
func NewOSCTarget() *OSCTarget {
	return &OSCTarget{conns: make(map[string]*net.UDPConn)}
}

func (t *OSCTarget) Type() mapping.OutputType { return mapping.OutputOSC }

func (t *OSCTarget) Execute(action mapping.OutputAction, value float64) error {
	host := action.Host
	if host == "" {
		host = "localhost"
	}
	port := action.Port
	if port == 0 {
		port = 7000
	}

	conn, err := t.getConn(host, port)
	if err != nil {
		return err
	}

	msg := buildOSCMessage(action.Address, float32(value))
	_, err = conn.Write(msg)
	return err
}

func (t *OSCTarget) getConn(host string, port int) (*net.UDPConn, error) {
	key := fmt.Sprintf("%s:%d", host, port)

	t.mu.Lock()
	defer t.mu.Unlock()

	if c, ok := t.conns[key]; ok {
		return c, nil
	}

	addr, err := net.ResolveUDPAddr("udp", key)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}
	t.conns[key] = conn
	return conn, nil
}

func (t *OSCTarget) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, c := range t.conns {
		c.Close()
	}
	t.conns = nil
	return nil
}

// buildOSCMessage creates an OSC message with a single float32 argument.
func buildOSCMessage(address string, value float32) []byte {
	return buildOSCBundle(address, ",f", func(buf []byte) []byte {
		var fb [4]byte
		binary.BigEndian.PutUint32(fb[:], math.Float32bits(value))
		return append(buf, fb[:]...)
	})
}

// buildOSCMessageInt creates an OSC message with a single int32 argument.
func buildOSCMessageInt(address string, value int32) []byte {
	return buildOSCBundle(address, ",i", func(buf []byte) []byte {
		var ib [4]byte
		binary.BigEndian.PutUint32(ib[:], uint32(value))
		return append(buf, ib[:]...)
	})
}

// buildOSCMessageString creates an OSC message with a single string argument.
func buildOSCMessageString(address string, value string) []byte {
	return buildOSCBundle(address, ",s", func(buf []byte) []byte {
		return append(buf, oscString(value)...)
	})
}

// buildOSCBundle constructs an OSC message with address, type tag, and argument writer.
func buildOSCBundle(address string, typeTag string, writeArgs func([]byte) []byte) []byte {
	var buf []byte
	buf = append(buf, oscString(address)...)
	buf = append(buf, oscString(typeTag)...)
	buf = writeArgs(buf)
	return buf
}

// oscString pads a string to a 4-byte boundary with null terminators.
func oscString(s string) []byte {
	b := []byte(s)
	b = append(b, 0) // null terminator
	for len(b)%4 != 0 {
		b = append(b, 0)
	}
	return b
}
