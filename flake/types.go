package flake

import "math/big"

// Generator is the primary interface for flake ID generation
type Generator interface {
	// Epoch returns the epoch used for overt-flake identifers
	Epoch() int64
	// HardwareID returns the hardware identifier used for overt-flake identifiers
	HardwareID() HardwareID
	// ProcessID returns the process id hosting the overt-flake Generator
	ProcessID() int

	// Get the size, in bytes, of IDs created by the generator
	IDSize() int

	// LastAllocatedTime is the last Unix Epoch value that one or more ids
	// are known to have been generated
	LastAllocatedTime() int64

	// Generate generates count overt-flake identifiers
	Generate(count int) ([]byte, error)

	// GenerateAsStream allocates and returns ids in chunks (based on the size of buffer) via a callback
	GenerateAsStream(count int, buffer []byte, callback func(int, []byte) error) (totalAllocated int, err error)
}

// HardwareID is an alias for []byte
type HardwareID []byte

// HardwareIDProvider is a provider that generates a hardware identifier for
// use by an overt-flake Generator
type HardwareIDProvider interface {
	GetHardwareID(byteSize int) ([]byte, error)
}

// OvertFlakeID is an interface that provides access to the components and
// alternate representations of an overt-flake identifier
type OvertFlakeID interface {
	// Timestamp is when the ID was generated, and is the # of milliseconds since
	// the generator Epoch
	Timestamp() uint64
	// SequenceID represents the Nth value created during a time interval (0 for the 1st)
	SequenceID() uint16

	// HardwareID is the HardwareID assigned by the generator
	HardwareID() HardwareID
	// ProcessID is the processID assigned by the generator
	ProcessID() uint16
	// MachineID is the uint64 representation of HardwareID and ProcessID and is == Lower()
	MachineID() uint64

	// Upper is the upper (most-signficant) bytes of the id represented as a uint64
	Upper() uint64
	// Lower is the lower (least-signficant) bytes of the id represented as a uint64
	Lower() uint64

	// Bytes is the []byte representation of the ID
	Bytes() []byte

	// ToBigInt converts the ID to a *big.Int
	ToBigInt() *big.Int

	// String returns the big.Int string representation of the ID
	String() string
}
