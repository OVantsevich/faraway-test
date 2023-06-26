package protocol

import (
	"net"
	"time"
)

//go:generate mockery --name=MockNetConn --case=underscore --output=./mocks
type MockNetConn interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	SetDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}
