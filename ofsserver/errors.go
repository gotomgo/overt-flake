package ofsserver

import (
	"errors"
	"fmt"
)

var (
	// ErrAuthRequired occurs after a client connection if the first 4 bytes received is not 0xFFFFFFnn
	// and the server requires authorization
	ErrAuthRequired = errors.New("Client authentication is required")
	// ErrInvalidAuth occurs when the client authentication is incorrect
	ErrInvalidAuth = errors.New("Invalid Credentials")
	// ErrInvalidReauthentication occurs when the client sends a authentication header after
	// the client has already authenticated
	ErrInvalidReauthentication = errors.New("Client is attempting unexpected re-authentication")
	// ErrShortWrite occurs when the server writes less bytes than it expected to write and it is
	// considered an error
	ErrShortWrite = errors.New("Expecting to write more bytes than were actually written")
)

// CreateArgumentNilError creates a custom form ErrArgumentNil with the argument
// name included in the error message
func CreateArgumentNilError(paramName string) error {
	return fmt.Errorf("The argument '%s' cannot be nil", paramName)
}

// CreateBadArgumentError creates a custom form ErrArgumentNil with the argument
// name included in the error message
func CreateBadArgumentError(paramName, message string, args ...interface{}) error {
	return fmt.Errorf("The value of the argument '%s' is invalid: %s", paramName, fmt.Sprintf(message, args))
}
