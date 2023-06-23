package protocol

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"hash"
	"math"
	"math/big"
	"math/rand"
	"net"
	"time"
)

// PowError represents an error encountered during Proof of Work.
// It must be processed within this connection.
type PowError struct {
	msg string
}

func (e *PowError) Error() string { return e.msg }

// challenge represents a Proof of Work challenge.
type challenge struct {
	pow *ProofOfWork

	data []byte
}

// newChallenge creates a new challenge.
func (pow *ProofOfWork) newChallenge(data []byte) *challenge {
	return &challenge{
		pow: pow,
		data: bytes.Join(
			[][]byte{
				data,
				i32tob(pow.rand.Uint32()),
			},
			[]byte{},
		),
	}
}

// marshal converts the challenge to a byte slice.
func (c *challenge) marshal() []byte {
	return bytes.Join(
		[][]byte{
			c.data,
		},
		[]byte{},
	)
}

// unmarshalChallenge converts a byte slice to a challenge.
func (pow *ProofOfWork) unmarshalChallenge(data []byte) *challenge {
	return &challenge{
		pow:  pow,
		data: data[:],
	}
}

// nonceLen - len of nonce variable in bytes
const nonceLen = 4

// response represents a Proof of Work response.
type response struct {
	pow *ProofOfWork

	hash  []byte
	nonce uint32
}

// newResponse creates a new response.
func (pow *ProofOfWork) newResponse(hash []byte, nonce uint32) *response {
	return &response{
		hash:  hash,
		nonce: nonce,
		pow:   pow,
	}
}

// marshal converts the response to a byte slice.
func (r *response) marshal() []byte {
	return bytes.Join(
		[][]byte{
			r.hash[:],
			i32tob(r.nonce),
		},
		[]byte{},
	)
}

// unmarshalResponse converts a byte slice to a response.
func (pow *ProofOfWork) unmarshalResponse(data []byte) (*response, error) {
	if len(data) != pow.hash.Size()+nonceLen {
		return nil, &PowError{"response is malformed"}
	}

	return &response{
		hash:  data[0:pow.hash.Size()],
		nonce: btoi32(data[pow.hash.Size():]),
		pow:   pow,
	}, nil
}

// randu32 represents an uint32 random number generator.
type randu32 interface {
	Uint32() uint32
}

// ProofOfWork represents the Proof of Work algorithm configuration.
type ProofOfWork struct {
	// targetBits the target number of requested leading zeros in the hash
	targetBits uint8

	// target is another name for the requirements described in the previous section.
	//We use big integer because of the way the hash is compared to the target: we convert the hash to a big integer and check if it is less than the target.
	//In the NewProofOfWork function, we initialize big.Int to 1 and then shift it by 256-targetBits bits.
	//256 is the length of the SHA-256 hash in bits, and we will use this hashing algorithm. Hex representation of target:
	target *big.Int

	// readTimeout - server timeout for waiting for a response
	readTimeout time.Duration

	// randu32 - uint32 random number generator
	rand randu32

	// hash - hash of PoW algorithm
	hash hash.Hash
}

// NewProofOfWork creates a new Proof of Work configuration.
func NewProofOfWork(targetBits uint8, hash hash.Hash, rand randu32, readTimeout time.Duration) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, 256-uint(targetBits))
	return &ProofOfWork{target: target, targetBits: targetBits, hash: hash, rand: rand, readTimeout: readTimeout}
}

// ChallengeResponse performs the Proof of Work challenge-response protocol.
func (pow *ProofOfWork) ChallengeResponse(conn net.Conn, data []byte) error {
	chal := pow.newChallenge(data)

	_, err := conn.Write(chal.marshal())
	if err != nil {
		return fmt.Errorf("ChallengeResponse: Write error: %v", err)
	}

	if pow.readTimeout != 0 {
		err = conn.SetReadDeadline(time.Now().Add(pow.readTimeout))
		if err != nil {
			return fmt.Errorf("ChallengeResponse: SetReadDeadline error: %v", err)
		}
	}

	rData := make([]byte, pow.hash.Size()+nonceLen)
	_, err = conn.Read(rData)
	if err != nil {
		return fmt.Errorf("ChallengeResponse: Read error: %v", err)
	}

	resp, err := pow.unmarshalResponse(rData)
	if err != nil {
		return err
	}

	if pow.validate(resp) && pow.compare(chal, resp) {
		return nil
	}
	return &PowError{"response is not valid"}
}

// SolveChallenge performs the Proof of Work challenge-solving protocol.
func SolveChallenge(conn net.Conn, targetBits uint8) error {
	s3 := rand.NewSource(time.Now().UnixNano())
	r3 := rand.New(s3)

	target := big.NewInt(1)
	target.Lsh(target, 256-uint(targetBits))

	pow := &ProofOfWork{
		target:      target,
		targetBits:  targetBits,
		readTimeout: 0,
		rand:        r3,
		hash:        sha256.New(),
	}

	return pow.SolveChallenge(conn)
}

// SolveChallenge performs the Proof of Work challenge-solving protocol.
func (pow *ProofOfWork) SolveChallenge(conn net.Conn) error {
	var data []byte
	_, err := conn.Read(data)
	if err != nil {
		return fmt.Errorf("SolveChallenge: Read error: %v", err)
	}
	chal := pow.unmarshalChallenge(data)

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
	if err != nil {
		return fmt.Errorf("SolveChallenge: Write error: %v", err)
	}

	return nil
}

// computeHash calculates the hash value for the given data and nonce.
func (pow *ProofOfWork) computeHash(data []byte, nonce uint32) []byte {
	return pow.hash.Sum(
		bytes.Join(
			[][]byte{
				i32tob(nonce),
				data,
			},
			[]byte{},
		))
}

// validate checks if the response hash is less than the target value.
func (pow *ProofOfWork) validate(r *response) bool {
	var hashInt big.Int

	hashInt.SetBytes(r.hash[:])
	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}

// compare checks if the newly computed hash is equal to the response hash.
func (pow *ProofOfWork) compare(c *challenge, r *response) bool {
	newHash := pow.computeHash(c.data, r.nonce)

	isEqual := bytes.Compare(newHash, r.hash) == 0

	return isEqual
}

// i32tob converts a uint32 value to a byte slice.
func i32tob(val uint32) []byte {
	r := make([]byte, 4)
	for i := uint32(0); i < 4; i++ {
		r[i] = byte((val >> (8 * i)) & 0xff)
	}
	return r
}

// btoi32 converts a byte slice to a uint32 value.
func btoi32(val []byte) uint32 {
	r := uint32(0)
	for i := uint32(0); i < 4; i++ {
		r |= uint32(val[i]) << (8 * i)
	}
	return r
}
