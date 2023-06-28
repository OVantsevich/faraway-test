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

type Client struct {
	conn net.Conn
	// Challenge-response protocol implementation
	crProto clientChallengeResponse
}

func NewClient(conn net.Conn) (*Client, error) {
	c := &Client{conn: conn}

	syn, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt8))
	if err != nil {
		return nil, fmt.Errorf("NewClient - Int: %v", err)
	}

	err = binary.Write(c.conn, binary.LittleEndian, int16(syn.Int64()))
	if err != nil {
		return nil, fmt.Errorf("NewClient - Int: %v", err)
	}

	var ack int32
	err = binary.Read(c.conn, binary.LittleEndian, &ack)
	if err != nil {
		return nil, fmt.Errorf("NewClient - Read: %v", err)
	}

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

// GetQuote - get quote
func (c *Client) GetQuote() (string, error) {
	_, err := c.conn.Write([]byte(fmt.Sprint("GetQuote", "\n")))
	if err != nil {
		return "", fmt.Errorf("GetQuote - Write: %v", err)
	}

	if c.crProto != nil {
		err := c.crProto.SolveChallenge(c.conn)
		if err != nil {
			return "", fmt.Errorf("GetQuote - SolveChallenge: %v", err)
		}
	}

	quote, err := bufio.NewReader(c.conn).ReadSlice('\n')
	if err != nil {
		return "", fmt.Errorf("GetQuote - ReadSlice: %v", err)
	}
	return string(quote), nil
}
