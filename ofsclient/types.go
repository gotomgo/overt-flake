package ofsclient

import (
	"math/big"

	"github.com/gotomgo/overt-flake/flake"
)

// Client represents the api provided by an ofsserver client
type Client interface {
	// GenerateIDBytes generates count ids in []byte form
	GenerateIDBytes(count int) (ids []byte, err error)

	// GenerateIDAsBigInt generates a single ID as a *big.Int
	GenerateIDAsBigInt() (*big.Int, error)
	// GenerateIDsAsBigInt generates count IDs in the form of []big.Int
	GenerateIDsAsBigInt(count int) (bigInts []big.Int, err error)

	// GenerateIDsAsFlake generates count IDs in the form of []flake.OvertFlakeID
	GenerateIDsAsFlake(count int) (flakes []flake.OvertFlakeID, err error)

	// GenerateIDAsFlake generates a single ID in the form of *flake.OvertFlakeID
	GenerateIDAsFlake() (flake *flake.OvertFlakeID, err error)

	// Close closes any open connection. If any generate call is made the connection
	// will be re-established
	Close()
}
