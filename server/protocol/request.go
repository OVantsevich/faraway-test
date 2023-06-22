// Package protocol provides client and server for communication using protocol
package protocol

// method of Request in protocol, represents type of Request
type method int

const (
	//syn         				// from client to server - request new challenge from server
	//ack         				// from server to client - message with challenge for client
	getQuote method = 10 * iota // request for Quote from server
)

type Request struct {
	Method method
}
