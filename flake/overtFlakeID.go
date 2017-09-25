package flake

import (
	"encoding/binary"
	"math/big"
)

// overtflakeID is a wrapper around the bytes generated for an overt-flake identifer
// 	- it implementst the OvertFlakeID interface
type overtFlakeID struct {
	idBytes []byte
}

// NewOvertFlakeID creates an instance of overtFlakeID which implements OvertFlakeID
func NewOvertFlakeID(id []byte) OvertFlakeID {
	return &overtFlakeID{
		idBytes: id,
	}
}

// Timestamp is when the ID was generated, and is the # of milliseconds since
// the generator Epoch
func (id *overtFlakeID) Timestamp() uint64 {
	return id.Upper() >> 16
}

// SequenceID represents the Nth value created during a time interval
// (0 if the 1st interval generated)
func (id *overtFlakeID) SequenceID() uint16 {
	return uint16(id.Upper() & 0xFFFF)
}

// HardwareID is the HardwareID assigned by the generator
func (id *overtFlakeID) HardwareID() HardwareID {
	return id.idBytes[8:14]
}

// ProcessID is the processID assigned by the generator
func (id *overtFlakeID) ProcessID() uint16 {
	return uint16(id.Lower() & 0xFFFF)
}

// MachineID is the uint64 representation of HardwareID and ProcessID and is == Lower()
func (id *overtFlakeID) MachineID() uint64 {
	return id.Lower()
}

// Upper is the upper (most-signficant) bytes of the id represented as a uint64
func (id *overtFlakeID) Upper() uint64 {
	return binary.BigEndian.Uint64(id.idBytes[0:8])
}

// Lower is the lower (least-signficant) bytes of the id represented as a uint64
func (id *overtFlakeID) Lower() uint64 {
	return binary.BigEndian.Uint64(id.idBytes[8:16])
}

// Bytes is the []byte representation of the ID
func (id *overtFlakeID) Bytes() []byte {
	return id.idBytes
}

// ToBigInt converts the ID to a *big.Int
func (id *overtFlakeID) ToBigInt() *big.Int {
	i := big.NewInt(0)
	i.SetBytes(id.idBytes)
	return i
}

// String returns the big.Int string representation of the ID
func (id *overtFlakeID) String() string {
	return id.ToBigInt().String()
}
