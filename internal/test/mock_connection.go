package test

import (
	"bytes"
	"net"
	"time"
)

type MockConn struct {
	readBuf  bytes.Buffer
	writeBuf chan []byte
}

func NewMockConn(data []byte) *MockConn {
	return &MockConn{
		readBuf:  *bytes.NewBuffer(data),
		writeBuf: make(chan []byte),
	}
}

func (c *MockConn) Read(b []byte) (int, error) {
	return c.readBuf.Read(b)
}

func (c *MockConn) Write(b []byte) (int, error) {
	c.writeBuf <- b
	return len(b), nil
}

func (c *MockConn) Close() error                       { return nil }
func (c *MockConn) LocalAddr() net.Addr                { return &net.TCPAddr{} }
func (c *MockConn) RemoteAddr() net.Addr               { return &net.TCPAddr{} }
func (c *MockConn) SetDeadline(t time.Time) error      { return nil }
func (c *MockConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *MockConn) SetWriteDeadline(t time.Time) error { return nil }

func (c *MockConn) ReadSent(buf []byte, timeout time.Duration) bool {
	timer := time.NewTimer(timeout)
	for {
		select {
		case data := <-c.writeBuf:
			if len(data) == len(buf) {
				copy(buf, data)
				return true
			}
		case <-timer.C:
			return false
		}
	}
}
