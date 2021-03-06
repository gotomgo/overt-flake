package ofsclient

import (
	"errors"
	"fmt"
)

var (
	// ErrNoServers indicates that the client was not provided with any server addresses
	ErrNoServers = errors.New("no server (network addresses) were provided for client connections")
	// ErrAuthTokenTooLarge indicates that client auth token exceeds the max length of 255 bytes
	ErrAuthTokenTooLarge = errors.New("auth token cannot exceed 255 bytes in length")
	// ErrShortRead indicates that the client read less bytes than it expected and it was
	// considered an error
	ErrShortRead = errors.New("Read less bytes than expected")
	// ErrNoServerConnection indicates that a server connection could not be established
	ErrNoServerConnection = errors.New("No ofsrvr connection could be established")
)

// CreateBadArgumentError creates a custom form ErrArgumentNil with the argument
// name included in the error message
func CreateBadArgumentError(paramName, message string, args ...interface{}) error {
	return fmt.Errorf("The value of the argument '%s' is invalid: %s", paramName, fmt.Sprintf(message, args...))
}
