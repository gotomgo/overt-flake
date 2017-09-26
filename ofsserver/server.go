package ofsserver

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/gotomgo/overt-flake/flake"
)

var (
	ErrAuthRequired   = errors.New("Client authentication is required")
	ErrInvalidRequest = errors.New("Invalid request")
	ErrInvalidAuth    = errors.New("Invalid Credentials")
	ErrShortWrite     = errors.New("Expecting to write more bytes than were actually written")
)

// OvertFlakeServer is a simple flake ID server based on NOEQD
type OvertFlakeServer struct {
	authToken string
	listener  net.Listener
	generator flake.Generator
	ipAddr    string
}

// NewOvertFlakeServer creates an instance of OvertFlakeServer
func NewOvertFlakeServer(generator flake.Generator, ipAddr, authToken string) *OvertFlakeServer {
	return &OvertFlakeServer{
		ipAddr:    ipAddr,
		generator: generator,
		authToken: authToken,
	}
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
	// if an authToken is specified then clients must send an auth sequence
	// FF FF FF n {n bytes} where {n bytes} is the client value for the auth token
	if server.authToken != "" {
		server.authenticateClient(reader)
	}

	// buffer for 16 IDs at a time
	buffer := make([]byte, 16*flake.OvertFlakeIDLength)

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
			return ErrInvalidRequest
		}

		_, err = server.generator.GenerateAsStream(int(count), buffer, func(allocated int, ids []byte) error {
			var bytesWritten int

			totalBytes := flake.OvertFlakeIDLength * allocated

			if allocated == (len(buffer) / flake.OvertFlakeIDLength) {
				bytesWritten, err = writer.Write(ids)
			} else {
				bytesWritten, err = writer.Write(ids[0:totalBytes])
			}

			if (err == nil) && (bytesWritten != totalBytes) {
				fmt.Printf("The bytesWritten, %d, was expected to be %d\n", bytesWritten, totalBytes)
				return ErrShortWrite
			}

			return err
		})

		if err != nil {
			return err
		}
	}
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

	tokenBytes := make([]byte, authCmd&0xFF)
	_, err = io.ReadFull(reader, tokenBytes)
	if err != nil {
		return err
	}

	if string(tokenBytes) != server.authToken {
		return ErrInvalidAuth
	}

	return nil
}
