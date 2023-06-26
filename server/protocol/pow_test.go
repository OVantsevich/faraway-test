package protocol

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"math"
	"math/big"
	"math/rand"
	"net"
	"testing"
	"time"
)

func testSolver(pow *ProofOfWork) {
	conn, _ := net.Dial("tcp", "localhost:12345")
	data := make([]byte, sha256.Size+nonceLen)
	_, err := conn.Read(data)
	fmt.Println(err)
	var buf bytes.Buffer
	io.Copy(&buf, conn)
	fmt.Println(buf.String())
	chal, _ := pow.unmarshalChallenge(data)

	//	pow.SolveChallenge(conn)

	resp := pow.newResponse(nil, 0)
	for resp.nonce < math.MaxInt32 {
		resp.hash = pow.computeHash(chal.data, resp.nonce)
		var hashInt big.Int

		hashInt.SetBytes(resp.hash[:])

		if pow.validate(resp) {
			break
		} else {
			resp.nonce++
		}
	}

	conn.Write(resp.marshal())
}

func TestProofOfWork_ChallengeResponse(t *testing.T) {
	var targetBits byte = 15
	s3 := rand.NewSource(time.Now().UnixNano())
	r3 := rand.New(s3)

	target := big.NewInt(1)
	target.Lsh(target, 256-uint(targetBits))

	pow := &ProofOfWork{
		target:      target,
		targetBits:  targetBits,
		readTimeout: 0,
		rand:        r3,
	}
	fmt.Println(target)

	data := "FDAFAWEFA"

	l, err := net.Listen("tcp", "localhost:12345")
	require.NoError(t, err)

	go testSolver(pow)

	conn, err := l.Accept()
	require.NoError(t, err)

	err = pow.ChallengeResponse(conn, []byte(data))
	require.NoError(t, err)
}

func TestProofOfWork_computeHash(t *testing.T) {
	var targetBits byte = 0
	s3 := rand.NewSource(time.Now().UnixNano())
	r3 := rand.New(s3)

	target := big.NewInt(1)
	target.Lsh(target, 256-uint(targetBits))

	pow := &ProofOfWork{
		target:      target,
		targetBits:  targetBits,
		readTimeout: 0,
		rand:        r3,
	}

	v := i32tob(431453199)
	fmt.Println(string(v))
	fmt.Println(stoi32("19b7740f"))
	a := pow.computeHash([]byte("fcadfa"), 10)
	fmt.Printf("%x\n", a)
}
