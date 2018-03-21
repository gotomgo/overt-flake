package ofsclient

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
)

type client struct {
	mutex       sync.Mutex
	conn        net.Conn
	servers     []ServerEntry
	idSize      int
	serverIndex int
}

// NewClient creates an instance of client which implements Client
func NewClient(idSize int, servers []ServerEntry) (Client, error) {
	if len(servers) == 0 {
		return nil, ErrNoServers
	}

	for _, server := range servers {
		if len(server.Auth) > 255 {
			return nil, ErrAuthTokenTooLarge
		}
	}

	return &client{servers: servers, idSize: idSize}, nil
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
	// server entries are assumed to be priority ordered
	for index, server := range c.servers {
		// connect to server
		c.conn, err = net.Dial("tcp", server.Server)
		if err != nil {
			continue
		}

		// authenticate (as necessary)
		err = c.authenticate(server.Auth)
		// connected and auth'ed? excellent
		if err == nil {
			c.serverIndex = index
			return nil
		}
	}

	return ErrNoServerConnection
}

// authenticate writes an authentication header and auth token to the server IFF
// len(client.authToken) > 0
func (c *client) authenticate(authToken string) (err error) {
	// do we need to auth?
	if len(authToken) > 0 {
		// create the header 0xFFFFFFnn where nn is the length of the auth token
		authBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(authBytes[0:4], uint32(0xFFFFFF00|len(authToken)))

		// write the auth header
		_, err = c.conn.Write(authBytes)
		if err != nil {
			return
		}

		// write the auth token
		_, err = c.conn.Write([]byte(authToken))
		// @NOTE: if err != nil We should probably assume the token was bad and return ErrAuthFailed

		return
	}
	return
}

func (c *client) StreamIDBytes(count int, buffer []byte, callback func(int, []byte) error) (totalAllocated int, err error) {
	bufferCount := len(buffer) / c.idSize
	if bufferCount == 0 {
		return 0, CreateBadArgumentError("buffer", "The buffer is too small (%d bytes) to hold a single ID (%d bytes)", len(buffer), c.idSize)
	}

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

	for count > 0 {
		var readCount int

		// max # if ids we should read
		if count-bufferCount > 0 {
			readCount = bufferCount
		} else {
			readCount = count
		}

		binary.BigEndian.PutUint32(countBytes, uint32(readCount))

		// write the count
		_, err = c.conn.Write(countBytes)
		if err != nil {
			return
		}

		var bytesRead int
		// the number of bytes we expect to read
		expectedBytes := readCount * c.idSize

		// read the # of bytes for readCount ids
		bytesRead, err = io.ReadFull(c.conn, buffer[0:expectedBytes])
		if err != nil {
			return totalAllocated, err
		}

		// did we read all the byes we expected?
		if bytesRead != expectedBytes {
			return totalAllocated, ErrShortRead
		}

		// update our counts prior to callback
		totalAllocated += readCount
		count -= readCount

		// stream readCount ids back to caller
		err = callback(readCount, buffer)
		if err != nil {
			return
		}
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

	var bytesRead int

	// read the ids
	bytesRead, err = io.ReadFull(c.conn, ids)
	if err != nil {
		return
	}

	// did we read all the byes we expected?
	if bytesRead != expectedBytes {
		fmt.Printf("Expected %d bytes but only read %d for id.Size %d\n", expectedBytes, bytesRead, c.idSize)
		return nil, ErrShortRead
	}

	return
}

// GenerateID generates a single ID in []byte form
func (c *client) GenerateID() (id []byte, err error) {
	return c.GenerateIDBytes(1)
}
