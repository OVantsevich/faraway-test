package protocol

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func testSolver(t *testing.T, pow *ProofOfWork, address string) {
	conn, _ := net.Dial("tcp", address)
	chal, err := pow.readChallenge(bufio.NewReader(conn))
	require.NoError(t, err)

	resp := pow.newResponse(nil, 0)
	for resp.nonce < math.MaxInt32 {
		resp.hash = pow.computeHash(chal.data, resp.nonce)
		if pow.validate(resp) {
			break
		} else {
			resp.nonce++
		}
	}

	_, err = conn.Write(resp.marshal())
	require.NoError(t, err)
}

var letter = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func TestProofOfWork_ChallengeResponse(t *testing.T) {
	for targetBits := 0; targetBits < 23; targetBits++ {
		data := "127.0.0.1"
		pow := NewProofOfWork(uint8(targetBits), 0)

		address := fmt.Sprintf("localhost:123%d", targetBits)
		l, err := net.Listen("tcp", address)
		require.NoError(t, err)

		go testSolver(t, pow, address)

		conn, err := l.Accept()
		require.NoError(t, err)

		err = pow.ChallengeResponse(conn, []byte(data))
		require.NoError(t, err)
	}

	targetBits := 15
	for i := 0; i < 10; i++ {
		data := make([]byte, i)
		for j := range data {
			data[j] = letter[rand.Intn(len(letter))]
		}
		pow := NewProofOfWork(uint8(targetBits), 0)

		address := fmt.Sprintf("localhost:111%d", rand.Intn(99))
		l, err := net.Listen("tcp", address)
		require.NoError(t, err)

		go testSolver(t, pow, address)

		conn, err := l.Accept()
		require.NoError(t, err)

		err = pow.ChallengeResponse(conn, data)
		require.NoError(t, err)
	}

	data := []byte("127.0.0.1")
	readTimeOut := time.Nanosecond

	pow := NewProofOfWork(uint8(targetBits), readTimeOut)

	address := fmt.Sprintf("localhost:111%d", rand.Intn(99))
	l, err := net.Listen("tcp", address)
	require.NoError(t, err)

	go testSolver(t, pow, address)

	conn, err := l.Accept()
	require.NoError(t, err)

	err = pow.ChallengeResponse(conn, data)
	require.Error(t, err)

	readTimeOut = time.Second * 10

	pow = NewProofOfWork(uint8(targetBits), readTimeOut)

	address = fmt.Sprintf("localhost:111%d", rand.Intn(99))
	l, err = net.Listen("tcp", address)
	require.NoError(t, err)

	go testSolver(t, pow, address)

	conn, err = l.Accept()
	require.NoError(t, err)

	err = pow.ChallengeResponse(conn, data)
	require.NoError(t, err)
}

func TestProofOfWork_SolveChallenge(t *testing.T) {
	for targetBits := 0; targetBits < 23; targetBits++ {
		data := "127.0.0.1"
		pow := NewProofOfWork(uint8(targetBits), 0)

		address := fmt.Sprintf("localhost:333%d", targetBits)
		l, err := net.Listen("tcp", address)
		require.NoError(t, err)

		go func() {
			conn, _ := net.Dial("tcp", address)
			pow.SolveChallenge(conn)
		}()

		conn, err := l.Accept()
		require.NoError(t, err)

		err = pow.ChallengeResponse(conn, []byte(data))
		require.NoError(t, err)
	}

	targetBits := 15
	for i := 0; i < 10; i++ {
		data := make([]byte, i)
		for j := range data {
			data[j] = letter[rand.Intn(len(letter))]
		}
		pow := NewProofOfWork(uint8(targetBits), 0)

		address := fmt.Sprintf("localhost:333%d", rand.Intn(99))
		l, err := net.Listen("tcp", address)
		require.NoError(t, err)

		go func() {
			conn, _ := net.Dial("tcp", address)
			pow.SolveChallenge(conn)
		}()

		conn, err := l.Accept()
		require.NoError(t, err)

		err = pow.ChallengeResponse(conn, data)
		require.NoError(t, err)
	}

	data := []byte("127.0.0.1")
	readTimeOut := time.Nanosecond

	pow := NewProofOfWork(uint8(targetBits), readTimeOut)

	address := fmt.Sprintf("localhost:333%d", rand.Intn(99))
	l, err := net.Listen("tcp", address)
	require.NoError(t, err)

	go func() {
		conn, _ := net.Dial("tcp", address)
		pow.SolveChallenge(conn)
	}()

	conn, err := l.Accept()
	require.NoError(t, err)

	err = pow.ChallengeResponse(conn, data)
	require.Error(t, err)

	readTimeOut = time.Second * 10

	pow = NewProofOfWork(uint8(targetBits), readTimeOut)

	address = fmt.Sprintf("localhost:333%d", rand.Intn(99))
	l, err = net.Listen("tcp", address)
	require.NoError(t, err)

	go func() {
		innerConn, _ := net.Dial("tcp", address)
		pow.SolveChallenge(innerConn)
	}()

	conn, err = l.Accept()
	require.NoError(t, err)

	err = pow.ChallengeResponse(conn, data)
	require.NoError(t, err)
}

func TestSolveChallenge(t *testing.T) {
	for targetBits := 0; targetBits < 23; targetBits++ {
		data := "127.0.0.1"
		pow := NewProofOfWork(uint8(targetBits), 0)

		address := fmt.Sprintf("localhost:222%d", targetBits)
		l, err := net.Listen("tcp", address)
		require.NoError(t, err)

		go func() {
			innerConn, _ := net.Dial("tcp", address)
			SolveChallenge(innerConn, uint8(targetBits))
		}()

		conn, err := l.Accept()
		require.NoError(t, err)

		err = pow.ChallengeResponse(conn, []byte(data))
		require.NoError(t, err)
	}

	targetBits := 15
	for i := 0; i < 10; i++ {
		data := make([]byte, i)
		for j := range data {
			data[j] = letter[rand.Intn(len(letter))]
		}
		pow := NewProofOfWork(uint8(targetBits), 0)

		address := fmt.Sprintf("localhost:222%d", rand.Intn(99))
		l, err := net.Listen("tcp", address)
		require.NoError(t, err)

		go func() {
			innerConn, _ := net.Dial("tcp", address)
			SolveChallenge(innerConn, uint8(targetBits))
		}()

		conn, err := l.Accept()
		require.NoError(t, err)

		err = pow.ChallengeResponse(conn, data)
		require.NoError(t, err)
	}

	data := []byte("127.0.0.1")
	readTimeOut := time.Nanosecond

	pow := NewProofOfWork(uint8(targetBits), readTimeOut)

	address := fmt.Sprintf("localhost:222%d", rand.Intn(99))
	l, err := net.Listen("tcp", address)
	require.NoError(t, err)

	go func() {
		innerConn, _ := net.Dial("tcp", address)
		SolveChallenge(innerConn, uint8(targetBits))
	}()

	conn, err := l.Accept()
	require.NoError(t, err)

	err = pow.ChallengeResponse(conn, data)
	require.Error(t, err)

	readTimeOut = time.Second * 10

	pow = NewProofOfWork(uint8(targetBits), readTimeOut)

	address = fmt.Sprintf("localhost:222%d", rand.Intn(99))
	l, err = net.Listen("tcp", address)
	require.NoError(t, err)

	go func() {
		innerConn, _ := net.Dial("tcp", address)
		SolveChallenge(innerConn, uint8(targetBits))
	}()

	conn, err = l.Accept()
	require.NoError(t, err)

	err = pow.ChallengeResponse(conn, data)
	require.NoError(t, err)
}
