package ofsclient

import (
	"encoding/binary"
	"math/rand"
	"net"
	"sync"
)

type client struct {
	mutex     sync.Mutex
	conn      net.Conn
	servers   []string
	authToken string
	idSize    int
}

// NewClient creates an instance of client which implements Client
func NewClient(authToken string, idSize int, servers ...string) (Client, error) {
	if len(servers) == 0 {
		return nil, ErrNoServers
	}

	if len(authToken) > 255 {
		return nil, ErrAuthTokenTooLarge
	}

	return &client{authToken: authToken, servers: servers, idSize: idSize}, nil
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
		// @NOTE: if err != nil We should probably assume the token was bad and return ErrAuthFailed

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
	expectedBytes := count * c.idSize

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

// GenerateID generates a single ID in []byte form
func (c *client) GenerateID() (id []byte, err error) {
	return c.GenerateIDBytes(1)
}
