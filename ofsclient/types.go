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

	// Close closes any open connection. If any generate call is made the connection
	// will be re-established
	Close()
}

// OvertFlakeClient represents APIs more-centric to overt-flake identifiers
// and is an extension of Client
type OvertFlakeClient interface {
	Client

	// GenerateBigInt generates a single ID as a *big.Int
	GenerateBigInt() (*big.Int, error)
	// GenerateBigInts generates count IDs in the form of []big.Int
	GenerateBigInts(count int) (bigInts []big.Int, err error)

	// GenerateFlakes generates count IDs in the form of []flake.OvertFlakeID
	GenerateFlakes(count int) (flakes []flake.OvertFlakeID, err error)

	// GenerateFlake generates a single ID in the form of *flake.OvertFlakeID
	GenerateFlake() (flake *flake.OvertFlakeID, err error)
}
