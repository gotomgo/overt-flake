package ofsclient

import (
	"math/big"

	"github.com/gotomgo/overt-flake/flake"
)

type overtFlakeClient struct {
	client
}

// NewOvertFlakeClient creates an instance of overtFlakeClient which implements OvertFlakeClient
func NewOvertFlakeClient(authToken string, servers ...string) (OvertFlakeClient, error) {
	if len(servers) == 0 {
		return nil, ErrNoServers
	}

	if len(authToken) > 255 {
		return nil, ErrAuthTokenTooLarge
	}

	return &overtFlakeClient{client: client{authToken: authToken, servers: servers, idSize: flake.OvertFlakeIDLength}}, nil
}

// GenerateIDs generates count overt-flake identifiers in the form of []big.Int
func (c *overtFlakeClient) GenerateBigInts(count int) (bigInts []big.Int, err error) {
	ids, err := c.GenerateIDBytes(count)
	if err != nil {
		return
	}

	bigInts = make([]big.Int, count)

	// convert each 16 bytes (flake.OvertFlakeIDLength) into a big.Int
	for i := 0; i < count; i++ {
		offset := i * flake.OvertFlakeIDLength
		flakeID := flake.NewOvertFlakeID(ids[offset : offset+flake.OvertFlakeIDLength])
		bigInts[i] = *flakeID.ToBigInt()
	}

	return
}

// GenerateID generates a single overt-flake identifier in the form of a *big.Int
func (c *overtFlakeClient) GenerateBigInt() (*big.Int, error) {
	ids, err := c.GenerateBigInts(1)
	if len(ids) == 0 {
		return nil, err
	}
	return &ids[0], nil
}

// GenerateIDsAsFlake generates count overt-flake identifiers in the form of []flake.OvertFlakeID
func (c *overtFlakeClient) GenerateFlakes(count int) (flakes []flake.OvertFlakeID, err error) {
	ids, err := c.GenerateIDBytes(count)
	if err != nil {
		return
	}

	flakes = make([]flake.OvertFlakeID, count)

	// convert each 16 bytes (flake.OvertFlakeIDLength) into a Flake
	for i := 0; i < count; i++ {
		offset := i * flake.OvertFlakeIDLength
		flakes[i] = flake.NewOvertFlakeID(ids[offset : offset+flake.OvertFlakeIDLength])
	}

	return
}

// GenerateID generates a single overt-flake identifier in the form of a *big.Int
func (c *overtFlakeClient) GenerateFlake() (*flake.OvertFlakeID, error) {
	ids, err := c.GenerateFlakes(1)
	if len(ids) == 0 {
		return nil, err
	}
	return &ids[0], nil
}
