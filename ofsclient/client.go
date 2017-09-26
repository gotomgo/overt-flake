package ofsclient

import (
	"encoding/binary"
	"math/big"
	"math/rand"
	"net"
	"sync"

	"github.com/gotomgo/overt-flake/flake"
)

type client struct {
	mutex     sync.Mutex
	conn      net.Conn
	servers   []string
	authToken string
}

// NewClient creates an instance of client which implements Client
func NewClient(authToken string, servers ...string) (Client, error) {
	if len(servers) == 0 {
		return nil, ErrNoServers
	}

	if len(authToken) > 255 {
		return nil, ErrAuthTokenTooLarge
	}

	return &client{authToken: authToken, servers: servers}, nil
}

func (c *client) Close() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}

// connect selects an available server endpoint and establishes a connection
// and initiates authentication
func (c *client) connect() (err error) {
	// select a server
	n := rand.Intn(len(c.servers))

	// connect to server
	c.conn, err = net.Dial("tcp", c.servers[n])
	if err != nil {
		return
	}

	// authenticate (as necessary)
	return c.authenticate()
}

// authenticate writes an authentication header and auth token to the server IFF
// len(client.authToken) > 0
func (c *client) authenticate() (err error) {
	// do we need to auth?
	if len(c.authToken) > 0 {
		// create the header 0xFFFFFFnn where nn is the length of the auth token
		authBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(authBytes[0:4], uint32(0xFFFFFF00|len(c.authToken)))

		// write the auth header
		_, err = c.conn.Write(authBytes)
		if err != nil {
			return
		}

		// write the auth token
		_, err = c.conn.Write([]byte(c.authToken))
		// @NOTE: We should probably assume the token was bad and return ErrAuthFailed

		return
	}
	return
}

// GenerateIDBytes generates count number of overt-flake identifiers in raw/byte form
func (c *client) GenerateIDBytes(count int) (ids []byte, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// on exit, if there was an error, clear the connection
	defer func() {
		if err != nil {
			c.conn = nil
		}
	}()

	// if we don't have a connection establish one
	if c.conn == nil {
		err = c.connect()
		if err != nil {
			return
		}
	}

	// create the command header, which is the count of ids
	countBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(countBytes, uint32(count))

	// write the count
	_, err = c.conn.Write(countBytes)
	if err != nil {
		return
	}

	// the number of bytes we expect to read
	expectedBytes := count * flake.OvertFlakeIDLength

	// create a bytes buffer to accumulate the ids
	ids = make([]byte, expectedBytes)
	// read the ids
	bytesRead, err := c.conn.Read(ids)
	if err != nil {
		return
	}

	// did we read all the byes we expected?
	if bytesRead != expectedBytes {
		return nil, ErrShortRead
	}

	return
}

// GenerateIDs generates count overt-flake identifiers in the form of []big.Int
func (c *client) GenerateIDsAsBigInt(count int) (bigInts []big.Int, err error) {
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
func (c *client) GenerateIDAsBigInt() (*big.Int, error) {
	ids, err := c.GenerateIDsAsBigInt(1)
	if len(ids) == 0 {
		return nil, err
	}
	return &ids[0], nil
}

// GenerateIDsAsFlake generates count overt-flake identifiers in the form of []flake.OvertFlakeID
func (c *client) GenerateIDsAsFlake(count int) (flakes []flake.OvertFlakeID, err error) {
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
func (c *client) GenerateIDAsFlake() (*flake.OvertFlakeID, error) {
	ids, err := c.GenerateIDsAsFlake(1)
	if len(ids) == 0 {
		return nil, err
	}
	return &ids[0], nil
}
