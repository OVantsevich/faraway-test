package protocol

import (
	"bufio"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"net"
)

type clientChallengeResponse interface {
	SolveChallenge(net.Conn) error
}

// Client for interaction with protocol Quote server.
type Client struct {
	// Network connection associated with the client
	conn net.Conn
	// Challenge-response protocol implementation
	crProto clientChallengeResponse
}

// NewClient creates a new client instance with the given network connection.
func NewClient(conn net.Conn) (*Client, error) {
	c := &Client{conn: conn}

	// Generate a random SYN value
	syn, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt8))
	if err != nil {
		return nil, fmt.Errorf("NewClient - Int: %v", err)
	}

	// Send the SYN value to the server
	err = binary.Write(c.conn, binary.LittleEndian, int16(syn.Int64()))
	if err != nil {
		return nil, fmt.Errorf("NewClient - Int: %v", err)
	}

	var ack int32
	// Read the ACK value from the server
	err = binary.Read(c.conn, binary.LittleEndian, &ack)
	if err != nil {
		return nil, fmt.Errorf("NewClient - Read: %v", err)
	}

	// Check if the ACK value matches the SYN value, and if not, initialize the challenge-response protocol
	if int64(ack) != syn.Int64() {
		target := big.NewInt(1)
		target.Lsh(target, hashBitLen-uint(int64(ack)-syn.Int64()))

		pow := &ProofOfWork{
			target:     target,
			targetBits: uint8(int64(ack) - syn.Int64()),
		}
		c.crProto = pow
	}

	return c, nil
}

// GetQuote sends a request to the server to get a quote.
func (c *Client) GetQuote() (string, error) {
	// Send the request to the server
	_, err := c.conn.Write([]byte(fmt.Sprint("GetQuote", "\n")))
	if err != nil {
		return "", fmt.Errorf("GetQuote - Write: %v", err)
	}

	// Solve the challenge if the challenge-response protocol is implemented
	if c.crProto != nil {
		err = c.crProto.SolveChallenge(c.conn)
		if err != nil {
			return "", fmt.Errorf("GetQuote - SolveChallenge: %v", err)
		}
	}

	// Read the quote from the server
	quote, err := bufio.NewReader(c.conn).ReadSlice('\n')
	if err != nil {
		return "", fmt.Errorf("GetQuote - ReadSlice: %v", err)
	}
	return string(quote[:len(quote)-1]), nil
}
