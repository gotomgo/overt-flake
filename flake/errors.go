package flake

import "errors"

// ErrNoNetworkInterfaces occurs in the odd case where there are no network interfaces
var ErrNoNetworkInterfaces = errors.New("No network interfaces are available")

// ErrNoHardwareAddresses occurs in the odd case where there are network interfaces
// but none of them have hardware addresses
var ErrNoHardwareAddresses = errors.New("No Hardware Adresses are available")

// ErrInvalidSizeForHardwareAddress occurs when the request size of the hardwardID is
// incompatible with the hardwaredID provider
var ErrInvalidSizeForHardwareAddress = errors.New("The requested size of the hardware ID is not supported by the provider")

// ErrTooManyRequested occurs when the count passed to the Generate.Generate(int) func exceeds the maximum allowed
// The maximum allowed is generally related to the maximum sequence number
var ErrTooManyRequested = errors.New("The # of ids requested exceeds the maximum amount for 1 request")

// ErrTimeIsMovingBackwards occurs when the clock moves backwards putting us into a position where we could
// produce duplicate ids. That is obviously bad, and ID generation cannot resume until time catches up to
// where we were
var ErrTimeIsMovingBackwards = errors.New("time is moving backwards. Cannot resume until last time is reached")

// ErrBufferTooSmall occurs when the user passes in a buffer that is too small to fit a single overt-flake ID
var ErrBufferTooSmall = errors.New("the buffer is too small to hold an overt-flake ID")
