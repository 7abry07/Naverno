package test

import (
	"bytes"
	"net"
	"time"
)

type MockConn struct {
	readBuf  bytes.Buffer
	writeBuf bytes.Buffer
}

func NewMockConn(data []byte) *MockConn {
	return &MockConn{
		readBuf:  *bytes.NewBuffer(data),
		writeBuf: *bytes.NewBuffer([]byte{}),
	}
}

func (c *MockConn) Read(b []byte) (int, error) {
	return c.readBuf.Read(b)
}

func (c *MockConn) Write(b []byte) (int, error) {
	return c.writeBuf.Write(b)
}

func (c *MockConn) Close() error                       { return nil }
func (c *MockConn) LocalAddr() net.Addr                { return nil }
func (c *MockConn) RemoteAddr() net.Addr               { return nil }
func (c *MockConn) SetDeadline(t time.Time) error      { return nil }
func (c *MockConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *MockConn) SetWriteDeadline(t time.Time) error { return nil }

func (c *MockConn) ReadSent(b []byte) (int, error) {
	return c.writeBuf.Read(b)
}
