package ofsserver

import (
	"encoding/binary"
	"io"
	"log"
	"net"

	"github.com/gotomgo/overt-flake/flake"
)

const (
	// MaxAuthTokenLength is the maximum allowable length of an authentication token
	MaxAuthTokenLength = 255
)

// OvertFlakeServer is a simple flake ID server based on NOEQD
type OvertFlakeServer struct {
	authToken string
	listener  net.Listener
	generator flake.Generator
	ipAddr    string
}

// NewOvertFlakeServer creates an instance of OvertFlakeServer
func NewOvertFlakeServer(generator flake.Generator, ipAddr, authToken string) (*OvertFlakeServer, error) {
	if generator == nil {
		return nil, CreateArgumentNilError("generator")
	}

	if len(ipAddr) == 0 {
		return nil, CreateBadArgumentError("ipAddr", "The value cannot be empty")
	}

	if len(authToken) > MaxAuthTokenLength {
		return nil, CreateBadArgumentError("authToken", "The length of an auth token cannot exceed %d", MaxAuthTokenLength)
	}

	return &OvertFlakeServer{
		ipAddr:    ipAddr,
		generator: generator,
		authToken: authToken,
	}, nil
}

// Serve activates an OvertFlakeServer to accept connections and process requests
func (server *OvertFlakeServer) Serve() error {
	listener, err := net.Listen("tcp", server.ipAddr)
	if err != nil {
		return err
	}

	server.listener = listener

	return server.acceptAndServe()
}

func (server *OvertFlakeServer) acceptAndServe() error {
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			return err
		}

		go func() {
			defer conn.Close()

			err := server.serveClient(conn, conn)
			if err != io.EOF {
				log.Println(err)
			}
		}()
	}
}

func (server *OvertFlakeServer) serveClient(reader io.Reader, writer io.Writer) error {
	var hasAuthed bool

	// if an authToken is specified then clients must send an auth sequence
	// FF FF FF n {n bytes} where {n bytes} is the client value for the auth token
	if server.authToken != "" {
		err := server.authenticateClient(reader)
		if err != nil {
			return err
		}

		hasAuthed = true
	}

	idSize := server.generator.IDSize()

	// buffer for 16 IDs at a time
	buffer := make([]byte, 16*idSize)

	// a generate command is a 4-byte int (BigEndian) representing the # of ids
	// to be generated
	countBytes := make([]byte, 4)

	for {
		// Wait for 1 byte request
		_, err := io.ReadFull(reader, countBytes)
		if err != nil {
			return err
		}

		count := binary.BigEndian.Uint32(countBytes)

		if count == 0 {
			// 0 is not a valid ID count
			return CreateBadArgumentError("count", "must be > 0")
		}

		// if authorization is NOT required but the client thinks it is then
		// we end up here going WTF? So, just complete the authentication process
		// (basically a NOP) and move on. In this case where authentication has
		// already occured, then its a violation of protocol and we error out
		if (uint32(count) & 0xFFFFFF00) == 0xFFFFFF00 {
			if !hasAuthed {
				err = server.doAuth(reader, uint8(count&0xFF))
				if err != nil {
					return err
				}
				hasAuthed = true
			} else {
				return ErrInvalidReauthentication
			}
		}

		_, err = server.generator.GenerateAsStream(int(count), buffer, func(allocated int, ids []byte) error {
			var bytesWritten int

			totalBytes := idSize * allocated

			if allocated == (len(buffer) / idSize) {
				bytesWritten, err = writer.Write(ids)
			} else {
				bytesWritten, err = writer.Write(ids[0:totalBytes])
			}

			if (err == nil) && (bytesWritten != totalBytes) {
				return ErrShortWrite
			}

			return err
		})

		if err != nil {
			return err
		}
	}
}

func (server *OvertFlakeServer) doAuth(reader io.Reader, tokenCount uint8) (err error) {
	tokenBytes := make([]byte, tokenCount&0xFF)
	_, err = io.ReadFull(reader, tokenBytes)
	if err != nil {
		return err
	}

	if len(server.authToken) > 0 {
		if string(tokenBytes) != server.authToken {
			return ErrInvalidAuth
		}
	}

	return nil
}

func (server *OvertFlakeServer) authenticateClient(reader io.Reader) error {
	authBytes := make([]byte, 4)
	_, err := io.ReadFull(reader, authBytes)
	if err != nil {
		return err
	}

	authCmd := binary.BigEndian.Uint32(authBytes)
	if (authCmd & 0xFFFFFF00) != 0xFFFFFF00 {
		return ErrAuthRequired
	}

	return server.doAuth(reader, uint8(authCmd&0xFF))
}
