package ofsclient

import (
	"math/big"

	"github.com/gotomgo/overt-flake/flake"
)

// Client represents a basic, ID agnostic api that interacts with
// an ofsserver client
type Client interface {
	// GenerateIDBytes generates count ids in []byte form
	GenerateIDBytes(count int) (ids []byte, err error)

	// GenerateID generates a single ID in []byte form
	GenerateID() (id []byte, err error)

	// StreamIDBytes generates ids in chunks passing them back to the caller as they arrive, via buffer
	StreamIDBytes(count int, buffer []byte, callback func(int, []byte) error) (totalAllocated int, err error)

	// Close closes any open connection. If any generate call is made the connection
	// will be re-established
	Close()
}

// OvertFlakeClient represents APIs more-centric to overt-flake identifiers
// and is an extension of Client
type OvertFlakeClient interface {
	Client

	// StreamBigInts generates big.Int's in chunks passing them back to the caller as they arrive, via buffer
	StreamBigInts(count int, bigInts []big.Int, callback func(int, []big.Int) error) (totalAllocated int, err error)

	// GenerateBigInt generates a single ID as a *big.Int
	GenerateBigInt() (*big.Int, error)

	// GenerateBigInts generates count IDs in the form of []big.Int
	GenerateBigInts(count int) (bigInts []big.Int, err error)

	// StreamFlakes generates ids in chunks passing them back to the caller as they arrive, via buffer
	StreamFlakes(count int, flakes []flake.OvertFlakeID, callback func(int, []flake.OvertFlakeID) error) (totalAllocated int, err error)

	// GenerateFlakes generates count IDs in the form of []flake.OvertFlakeID
	GenerateFlakes(count int) (flakes []flake.OvertFlakeID, err error)

	// GenerateFlake generates a single ID in the form of *flake.OvertFlakeID
	GenerateFlake() (flake flake.OvertFlakeID, err error)
}
