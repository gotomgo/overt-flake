package ofsclient

import (
	"encoding/binary"
	"errors"
	"math/big"
	"math/rand"
	"net"
	"sync"

	"github.com/gotomgo/overt-flake/flake"
)

var (
	ErrNoAddrs      = errors.New("no network addresses provided")
	ErrInvalidToken = errors.New("auth token > 255 bytes in length")
)

type Client struct {
	mu    sync.Mutex
	cn    net.Conn
	addrs []string
	token string
}

func New(token string, addrs ...string) (*Client, error) {
	if len(addrs) == 0 {
		return nil, ErrNoAddrs
	}

	if len(token) > 255 {
		return nil, ErrInvalidToken
	}

	return &Client{token: token, addrs: addrs}, nil
}

func (c *Client) connect() (err error) {
	n := rand.Intn(len(c.addrs))
	c.cn, err = net.Dial("tcp", c.addrs[n])
	if err != nil {
		return
	}

	return c.auth()
}

func (c *Client) auth() (err error) {
	if c.token != "" {
		authBytes := make([]byte, 4+len(c.token))
		binary.BigEndian.PutUint32(authBytes[0:4], uint32(0xFFFFFF00|len(c.token)))
		copy(authBytes[4:], []byte(c.token))
		_, err = c.cn.Write(authBytes)
		return
	}
	return
}

func (c *Client) Gen(n int) (bigInts []big.Int, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	defer func() {
		if err != nil {
			c.cn = nil
		}
	}()

	if c.cn == nil {
		err = c.connect()
		if err != nil {
			return
		}
	}

	countBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(countBytes, uint32(n))

	_, err = c.cn.Write(countBytes)
	if err != nil {
		return
	}

	bigInts = make([]big.Int, n)

	ids := make([]byte, flake.OvertFlakeIDLength*n)
	err = binary.Read(c.cn, binary.BigEndian, ids)

	for i := 0; i < n; i++ {
		offset := i * flake.OvertFlakeIDLength
		flakeID := flake.NewOvertFlakeID(ids[offset : offset+flake.OvertFlakeIDLength])
		bigInts[i] = *flakeID.ToBigInt()
	}
	return
}

func (c *Client) GenOne() (*big.Int, error) {
	ids, err := c.Gen(1)
	if len(ids) == 0 {
		return nil, err
	}
	return &ids[0], nil
}
