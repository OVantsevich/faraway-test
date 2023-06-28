// Package protocol provides client and server for communication using protocol
package protocol

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"go.uber.org/zap"
)

type serverChallengeResponse interface {
	ChallengeResponse(conn net.Conn, data []byte) error // Method for generating a challenge response
	IncreaseComplexity()                                // Method to increase the complexity of the protocol
	DecreaseComplexity()                                // Method to decrease the complexity of the protocol
	GetComplexity() int                                 // Method to get the current complexity level of the protocol
	IsError(error) bool                                 // Method to check if an error is of type 'PowError'
}

type Request string

type Response string

// Handler function to be executed for incoming connections
type Handler func(Request) (Response, error)

// Server for handling tcp connection for Quote server
type Server struct {
	// Logger for logging server events
	logger *zap.SugaredLogger
	// Challenge-response protocol implementation
	crProto serverChallengeResponse
	// Timeout for the SYN operation
	synTimeout time.Duration
	// Handler function to be executed for incoming connections
	handler Handler
}

// NewServer creates a new instance of the server.
func NewServer(logger *zap.SugaredLogger, crProto serverChallengeResponse, synTimeout time.Duration, handler Handler) *Server {
	return &Server{logger: logger, crProto: crProto, synTimeout: synTimeout, handler: handler}
}

// newConn creates a new connection object associated with the server.
func (s *Server) newConn(rwc net.Conn) *conn {
	return &conn{server: s, rwc: rwc}
}

// Serve starts accepting and serving incoming connections.
//
//nolint:gomnd // const for retry needs no explanation
func (s *Server) Serve(l net.Listener) error {
	var tempDelay time.Duration

	for {
		rw, err := l.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && !ne.Timeout() {
				// Retry accepting connections with an increasing delay in case of non-timeout error
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

		// Handle the incoming connection in a separate goroutine
		c := s.newConn(rw)
		go c.serve()
	}
}

type conn struct {
	// Pointer to the server object
	server *Server
	// Underlying network connection associated with the connection
	rwc net.Conn
}

// serve is the main function for handling a connection.
func (c *conn) serve() {
	syn, err := c.syn()
	if err != nil {
		c.server.logger.Errorf("serve - syn: %v", err)
		return
	}

	// Retrieve the complexity level from the challenge-response protocol, if implemented
	var crComplexity int16
	if c.server.crProto != serverChallengeResponse(nil) {
		crComplexity = int16(c.server.crProto.GetComplexity())
	}
	err = c.ack(syn, crComplexity)
	if err != nil {
		c.server.logger.Errorf("serve - ack: %v", err)
		return
	}

	for {
		// Receive a message from the client
		req, err := bufio.NewReader(c.rwc).ReadSlice('\n')
		if err != nil {
			if err == io.EOF {
				return
			}
			c.server.logger.Errorf("serve - ReadSlice: %v", err)
			if err = c.rwc.Close(); err != nil {
				c.server.logger.Fatal(err)
			}
			return
		}

		// Perform challenge-response, if the protocol is implemented
		if c.server.crProto != serverChallengeResponse(nil) {
			err = c.server.crProto.ChallengeResponse(c.rwc, []byte(fmt.Sprint(c.rwc.RemoteAddr().String(), c.rwc.LocalAddr().String())))
			if err != nil {
				if !c.server.crProto.IsError(err) {
					c.server.logger.Errorf("serve - ChallengeResponse: %v. From = %s.", err, c.rwc.RemoteAddr().String())
				}
				if err = c.rwc.Close(); err != nil {
					c.server.logger.Fatal(err)
				}
				return
			}
		}

		// Call the server's handler function to handle the connection
		res, err := c.server.handler(Request(req))
		if err != nil {
			c.server.logger.Errorf("serve - handler: %v", err)
			if err = c.rwc.Close(); err != nil {
				c.server.logger.Fatal(err)
			}
			return
		}
		// Send Response to the client
		_, err = c.rwc.Write([]byte(fmt.Sprint(string(res), "\n")))
		if err != nil {
			c.server.logger.Errorf("serve - Write: %v", err)
			if err = c.rwc.Close(); err != nil {
				c.server.logger.Fatal(err)
			}
			return
		}
	}
}

// syn reads the SYN value from the connection.
func (c *conn) syn() (int16, error) {
	err := c.rwc.SetReadDeadline(time.Now().Add(c.server.synTimeout))
	if err != nil {
		return 0, fmt.Errorf("syn - Read: %v", err)
	}

	var syn int16
	err = binary.Read(c.rwc, binary.LittleEndian, &syn)
	if err != nil {
		return 0, fmt.Errorf("syn - Read: %v", err)
	}

	return syn, nil
}

// ack sends the ACK value (syn + pow) to the connection.
func (c *conn) ack(syn, pow int16) error {
	data := int32(syn + pow)
	err := binary.Write(c.rwc, binary.LittleEndian, data)
	return err
}
