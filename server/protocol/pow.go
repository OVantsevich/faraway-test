package protocol

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"net"
	"strconv"
	"time"
)

const (
	// eom - const representing end of message part of pow communication.
	eom = "EOM"
	// del - const representing delimiter between parts in pow messages.
	del = '|'
	// challengeParts - const representing number of challenge message parts in marshal/unmarshal.
	challengeParts = 3
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
	parts := [][]byte{
		i32tob(uint32(len(c.data))),
		c.data,
		[]byte(eom),
	}

	if len(parts) != challengeParts {
		panic("Invalid \"challengeParts\" const.")
	}

	return bytes.Join(
		parts,
		[]byte{del},
	)
}

// unmarshalChallenge converts a byte slice to a challenge.
func (pow *ProofOfWork) unmarshalChallenge(data []byte) (*challenge, error) {
	parts := bytes.Split(data, []byte{del})
	if len(parts) != challengeParts {
		return nil, &PowError{"challenge message is malformed"}
	}

	return &challenge{
		pow:  pow,
		data: parts[1],
	}, nil
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
			{'F'},
		},
		[]byte{'|'},
	)
}

// unmarshalResponse read data from bufio.Reader to a response.
func (pow *ProofOfWork) unmarshalResponse(reader bufio.Reader) (*response, error) {
	reader.
		parts := bytes.Split(data, []byte{'|'})
	if len(parts) != 2 {
		return nil, &PowError{"response is malformed"}
	}

	nonce, err := btoi32(parts[1])
	if err != nil {
		return nil, fmt.Errorf("unmarshalResponse: %v", err)
	}

	return &response{
		hash:  parts[0],
		nonce: nonce,
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
}

// NewProofOfWork creates a new Proof of Work configuration.
func NewProofOfWork(targetBits uint8, readTimeout time.Duration) *ProofOfWork {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	target := big.NewInt(1)
	target.Lsh(target, 256-uint(targetBits))
	return &ProofOfWork{target: target, targetBits: targetBits, rand: r, readTimeout: readTimeout}
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

	rData := make([]byte, sha256.Size+nonceLen)
	_, err = conn.Read(rData)
	if err != nil {
		return fmt.Errorf("ChallengeResponse: Read error: %v", err)
	}

	resp, err := pow.unmarshalResponse(rData)
	if err != nil {
		return err
	}

	if !pow.validate(resp) || !pow.compare(chal, resp) {
		fmt.Printf("%s\n", chal.data)
		fmt.Printf("%v\n", resp.nonce)
		fmt.Printf("%x\n", pow.computeHash(chal.data, resp.nonce))
		return &PowError{"response is not valid"}
	}
	return nil
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
	}

	return pow.SolveChallenge(conn)
}

// SolveChallenge performs the Proof of Work challenge-solving protocol.
func (pow *ProofOfWork) SolveChallenge(conn net.Conn) error {
	var data []byte
	_, err := conn.Read(data)
	netReader := bufio.NewReader(conn)
	netReader.read

	if err != nil {
		return fmt.Errorf("SolveChallenge: Read error: %v", err)
	}
	chal, err := pow.unmarshalChallenge(data)
	if err != nil {
		return fmt.Errorf("SolveChallenge: %v", err)
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
	hashInt.SetBytes(r.hash[:])
	return hashInt.Cmp(pow.target) == -1
}

// compare checks if the newly computed hash is equal to the response hash.
func (pow *ProofOfWork) compare(c *challenge, r *response) bool {
	newHash := pow.computeHash(c.data, r.nonce)
	return bytes.Compare(newHash, r.hash) == 0
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
		return 0, fmt.Errorf("i32tob - ParseUint: %v", err)
	}
	return uint32(val), nil
}

// stoi32 converts a string in hex format to a uint32 value.
func stoi32(hex string) (uint32, error) {
	val, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return 0, fmt.Errorf("i32tob - ParseUint: %v", err)
	}
	return uint32(val), nil
}
