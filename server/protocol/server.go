package protocol

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
)

type challengeResponse interface {
	ChallengeResponse(conn *bufio.ReadWriter, data []byte) error

	//IsCRError(error) bool
}

type conn struct {
	server *Server
	rwc    net.Conn
}

func (c *conn) serve() error {
	if c.server.crProto != challengeResponse(nil) {
		err := c.server.crProto.ChallengeResponse(bufio.NewReadWriter(bufio.NewReader(c.rwc), bufio.NewWriter(c.rwc)))

		if err != nil {
			if !c.server.crProto.IsCRError(err) {
				return fmt.Errorf("serve - ChallengeResponse: %v", err)
			}
			c.rwc.Close()
			return nil
		}
	}
}

func (c *conn) syn() error {
	c.rwc.SetReadDeadline()
}

func (c *conn) ack() error {

}

type Server struct {
	Addr string

	logger *zap.SugaredLogger

	crProto challengeResponse

	synTimeout time.Duration
}

func NewServer(addr string, logger *zap.SugaredLogger, crProto challengeResponse) *Server {
	return &Server{Addr: addr, logger: logger, crProto: crProto}
}

func (s *Server) newConn(rwc net.Conn) *conn {
	return &conn{server: s, rwc: rwc}
}

func (s *Server) Serve(l net.Listener) error {

	var tempDelay time.Duration

	for {
		rw, err := l.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && !ne.Timeout() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				s.logger.Errorf("protocol: Accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return err
		}
		c := s.newConn(rw)
		go c.serve()
	}
}
