package protocol

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"math"
	"math/big"
	"net"
	"strconv"
	"time"
)

const (
	// eom - const representing end of message part of pow communication.
	eom = "EOM"
	// del - const representing delimiter between parts in pow messages.
	del = '|'
	// hashBitLen - const representing len of sha checksum in bits.
	hashBitLen = 256
)

// PowError represents an error encountered during Proof of Work.
// It must be processed within this connection.
type PowError struct {
	msg string
}

func (e *PowError) Error() string { return e.msg }

// ProofOfWork represents the Proof of Work algorithm configuration.
type ProofOfWork struct {
	// targetBits the target number of requested leading zeros in the hash
	targetBits uint8

	// target is another name for the requirements described in the previous section.
	// We use big integer because of the way the hash is compared to the target: we convert the hash to a big integer and check if it is less than the target.
	// In the NewProofOfWork function, we initialize big.Int to 1 and then shift it by 256-targetBits bits.
	// 256 is the length of the SHA-256 hash in bits, and we will use this hashing algorithm. Hex representation of target:
	target *big.Int

	// readTimeout - server timeout for waiting for a response
	readTimeout time.Duration
}

// NewProofOfWork creates a new Proof of Work configuration.
func NewProofOfWork(targetBits uint8, readTimeout time.Duration) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, hashBitLen-uint(targetBits))
	return &ProofOfWork{target: target, targetBits: targetBits, readTimeout: readTimeout}
}

// ChallengeResponse performs the Proof of Work challenge-response protocol.
func (pow *ProofOfWork) ChallengeResponse(conn net.Conn, data []byte) error {
	rnd, err := rand.Int(rand.Reader, big.NewInt(math.MaxUint32))
	if err != nil {
		return fmt.Errorf("ChallengeResponse - Int: %v", err)
	}

	chal := pow.newChallenge(data, uint32(rnd.Int64()))

	_, err = conn.Write(chal.marshal())
	if err != nil {
		return fmt.Errorf("ChallengeResponse: Write error: %v", err)
	}

	if pow.readTimeout != 0 {
		err = conn.SetReadDeadline(time.Now().Add(pow.readTimeout))
		if err != nil {
			return fmt.Errorf("ChallengeResponse: SetReadDeadline error: %v", err)
		}
	}

	resp, err := pow.readResponse(bufio.NewReader(conn))
	if err != nil {
		return fmt.Errorf("ChallengeResponse - readResponse: %v", err)
	}

	if !pow.validate(resp) || !pow.compare(chal, resp) {
		return &PowError{"response is not valid"}
	}
	return nil
}

// SolveChallenge performs the Proof of Work challenge-solving protocol.
func SolveChallenge(conn net.Conn, targetBits uint8) error {
	target := big.NewInt(1)
	target.Lsh(target, hashBitLen-uint(targetBits))

	pow := &ProofOfWork{
		target:      target,
		targetBits:  targetBits,
		readTimeout: 0,
	}

	return pow.SolveChallenge(conn)
}

// SolveChallenge performs the Proof of Work challenge-solving protocol.
func (pow *ProofOfWork) SolveChallenge(conn net.Conn) error {
	chal, err := pow.readChallenge(bufio.NewReader(conn))
	if err != nil {
		return fmt.Errorf("SolveChallenge - readChallenge: %v", err)
	}

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

// challenge represents a Proof of Work challenge.
type challenge struct {
	pow *ProofOfWork

	data []byte
}

// newChallenge creates a new challenge.
func (pow *ProofOfWork) newChallenge(data []byte, rnd uint32) *challenge {
	return &challenge{
		pow: pow,
		data: bytes.Join(
			[][]byte{
				data,
				i32tob(rnd),
			},
			[]byte{},
		),
	}
}

// readChallenge read data from bufio.Reader to a challenge.
func (pow *ProofOfWork) readChallenge(reader *bufio.Reader) (*challenge, error) {
	data, err := reader.ReadBytes(del)
	if err != nil {
		return nil, fmt.Errorf("readChallenge - ReadSlice: %v", err)
	}
	data = data[:len(data)-1]

	return &challenge{
		pow:  pow,
		data: data,
	}, nil
}

// marshal converts the challenge to a byte slice.
func (c *challenge) marshal() []byte {
	parts := [][]byte{
		c.data,
		[]byte(eom),
	}
	return bytes.Join(
		parts,
		[]byte{del},
	)
}

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

// readResponse read data from bufio.Reader to a response.
func (pow *ProofOfWork) readResponse(reader *bufio.Reader) (*response, error) {
	hash := make([]byte, sha256.Size)
	_, err := io.ReadFull(reader, hash)
	if err != nil {
		return nil, fmt.Errorf("readResponse - ReadFull: %v", err)
	}
	_, err = reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("readResponse - ReadByte: %v", err)
	}

	nonce, err := reader.ReadBytes(del)
	if err != nil {
		return nil, fmt.Errorf("readResponse - ReadSlice: %v", err)
	}
	nonce = nonce[:len(nonce)-1]

	iNonce, err := btoi32(nonce)
	if err != nil {
		return nil, fmt.Errorf("readResponse - btoi32: %v", err)
	}

	return &response{
		hash:  hash,
		nonce: iNonce,
		pow:   pow,
	}, nil
}

// marshal converts the response to a byte slice.
func (r *response) marshal() []byte {
	parts := [][]byte{
		r.hash,
		i32tob(r.nonce),
		[]byte(eom),
	}
	return bytes.Join(
		parts,
		[]byte{del},
	)
}

// computeHash calculates the hash value for the given data and nonce.
func (pow *ProofOfWork) computeHash(data []byte, nonce uint32) []byte {
	checksum := sha256.Sum256(bytes.Join(
		[][]byte{
			i32tob(nonce),
			data,
		},
		[]byte{},
	))
	return checksum[:]
}

// validate checks if the response hash is less than the target value.
func (pow *ProofOfWork) validate(r *response) bool {
	var hashInt big.Int
	hashInt.SetBytes(r.hash)
	return hashInt.Cmp(pow.target) == -1
}

// compare checks if the newly computed hash is equal to the response hash.
func (pow *ProofOfWork) compare(c *challenge, r *response) bool {
	newHash := pow.computeHash(c.data, r.nonce)
	return bytes.Equal(newHash, r.hash)
}

// i32tob converts a uint32 value to a byte slice in hex format.
func i32tob(val uint32) []byte {
	hex := fmt.Sprintf("%x", val)
	return []byte(hex)
}

// btoi32 converts a byte slice in hex format to a uint32 value.
func btoi32(hex []byte) (uint32, error) {
	val, err := strconv.ParseUint(string(hex), 16, 32)
	if err != nil {
		return 0, fmt.Errorf("btoi32 - ParseUint: %v", err)
	}
	return uint32(val), nil
}
