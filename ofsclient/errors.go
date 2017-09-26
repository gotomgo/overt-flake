package ofsclient

import "errors"

var (
	// ErrNoServers indicates that the client was not provided with any server addresses
	ErrNoServers = errors.New("no server (network addresses) were provided for client connections")
	// ErrAuthTokenTooLarge indicates that client auth token exceeds the max length of 255 bytes
	ErrAuthTokenTooLarge = errors.New("auth token cannot exceed 255 bytes in length")
	// ErrShortRead indicates that the client read less bytes than it expected and it was
	// considered an error
	ErrShortRead = errors.New("Read less bytes than expected")
)
