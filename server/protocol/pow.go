package protocol

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
)

const hashSize = 64
const responseSize = 68

type PowError struct {
	msg string
}

func (e *PowError) Error() string { return e.msg }

type challenge struct {
	data   []byte
	target uint8
}

func newChallenge(rand uint32, data []byte, target uint8) *challenge {
	return &challenge{
		data: bytes.Join(
			[][]byte{
				data,
				i32tob(rand),
			},
			[]byte{},
		),
		target: target,
	}
}

func (c *challenge) marshal() []byte {
	return bytes.Join(
		[][]byte{
			c.data,
			{c.target},
		},
		[]byte{},
	)
}

func unmarshalChallenge(data []byte) *challenge {
	return &challenge{
		data:   data[:len(data)-1],
		target: data[len(data)-1],
	}
}

type response struct {
	hash  [hashSize]byte
	nonce uint32
}

func newResponse(hash [hashSize]byte, nonce uint32) *response {
	return &response{
		hash:  hash,
		nonce: nonce,
	}
}

func (c *response) marshal() []byte {
	return bytes.Join(
		[][]byte{
			c.hash[:],
			i32tob(c.nonce),
		},
		[]byte{},
	)
}

func unmarshalResponse(data []byte) (*response, error) {
	if len(data) != responseSize {
		return nil, &PowError{"response is malformed"}
	}

	return &response{
		hash:  [64]byte(data[0:hashSize]),
		nonce: btoi32(data[hashSiz e:]),
	}, nil
}

type ProofOfWork struct {
	target *big.Int
}

func NewProofOfWork(targetBits uint) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, 256-targetBits)
	return &ProofOfWork{target}
}

func (pow *ProofOfWork) ChallengeResponse(rwb *bufio.ReadWriter, data []byte) error {
	rwb.Read()
}

func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.Data,
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
}

func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Data)
	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Print("\n\n")

	return nonce, hash[:]
}

func i32tob(val uint32) []byte {
	r := make([]byte, 4)
	for i := uint32(0); i < 4; i++ {
		r[i] = byte((val >> (8 * i)) & 0xff)
	}
	return r
}

func btoi32(val []byte) uint32 {
	r := uint32(0)
	for i := uint32(0); i < 4; i++ {
		r |= uint32(val[i]) << (8 * i)
	}
	return r
}
