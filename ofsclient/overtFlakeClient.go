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

	serverEntrys := make([]ServerEntry, len(servers))
	for index, server := range servers {
		serverEntrys[index] = ServerEntry{Server: server, Auth: authToken}
	}

	return &overtFlakeClient{client: client{servers: serverEntrys, idSize: flake.OvertFlakeIDLength}}, nil
}

// NewOvertFlakeClientWithConfig creates an instance of overtFlakeClient which implements OvertFlakeClient
func NewOvertFlakeClientWithConfig(config *Config) (OvertFlakeClient, error) {
	if len(config.Servers) == 0 {
		return nil, ErrNoServers
	}

	for _, server := range config.Servers {
		if len(server.Auth) > 255 {
			return nil, ErrAuthTokenTooLarge
		}
	}

	return &overtFlakeClient{client: client{servers: config.Servers, idSize: flake.OvertFlakeIDLength}}, nil
}

// StreamFlakes generates ids in chunks passing them back to the caller as they arrive, via buffer
func (c *overtFlakeClient) StreamBigInts(count int, bigInts []big.Int,
	callback func(int, []big.Int) error) (totalAllocated int, err error) {

	// create a byte buffer with a length that corresponds to the flakes buffer provided by caller
	buffer := make([]byte, len(bigInts)*flake.OvertFlakeIDLength)
	totalAllocated, err = c.StreamIDBytes(count, buffer, func(count int, ids []byte) error {
		// as the id bytes stream in, convert them to flakes
		for i := 0; i < count; i++ {
			index := i * flake.OvertFlakeIDLength
			bigInts[i] = *flake.NewOvertFlakeID(buffer[index : index+flake.OvertFlakeIDLength]).ToBigInt()
		}

		// send the flakes to our caller
		return callback(count, bigInts)
	})

	return
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

// StreamFlakes generates ids in chunks passing them back to the caller as they arrive, via buffer
func (c *overtFlakeClient) StreamFlakes(count int, flakes []flake.OvertFlakeID,
	callback func(int, []flake.OvertFlakeID) error) (totalAllocated int, err error) {

	// create a byte buffer with a length that corresponds to the flakes buffer provided by caller
	buffer := make([]byte, len(flakes)*flake.OvertFlakeIDLength)
	totalAllocated, err = c.StreamIDBytes(count, buffer, func(count int, ids []byte) error {
		// as the id bytes stream in, convert them to flakes
		for i := 0; i < count; i++ {
			index := i * flake.OvertFlakeIDLength
			flakes[i] = flake.NewOvertFlakeID(buffer[index : index+flake.OvertFlakeIDLength])
		}

		// send the flakes to our caller
		return callback(count, flakes)
	})

	return
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
func (c *overtFlakeClient) GenerateFlake() (flake.OvertFlakeID, error) {
	ids, err := c.GenerateFlakes(1)
	if len(ids) == 0 {
		return nil, err
	}
	return ids[0], nil
}
