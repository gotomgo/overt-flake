package flake

import (
	"encoding/binary"
)

const (
	// SequenceBits53 is the # of bits used for 53-MSB sequence #'s
	SequenceBits53 uint64 = 12
	// MSBMask53 represents 2^53-1
	MSBMask53 = 0x1FFFFFFFFFFFFF
)

//  ---------------------------------------------------------------------------
//  Layout - Big Endian
//  ---------------------------------------------------------------------------
//
//  [0:6]   48 bits | Upper 48 bits of timestamp (milliseconds since the epoch)
//  [6:8]   16 bits | a per-interval sequence # (interval == 1 millisecond)
//  [8:14]  48 bits | a hardware id
//  [14:16] 16 bits | process ID
//
//  ---------------------------------------------------------------------------
//  | 0 | 1  | 2 | 3 | 4 | 5 |  6  |  7  |  8  | 9 | A | B | C | D |  E  |  F  |
//  ---------------------------------------------------------------------------
//  |   00   |    43 bits    |  10 bits  |        48 bits          |  16 bits  |
//  ---------------------------------------------------------------------------
//  | unused |   timestamp   |  sequence  |       HardwareID        | ProcID    |
//  ---------------------------------------------------------------------------
//  Notes
//  ---------------------------------------------------------------------------
//  The time bits are the most significant bits because they have the primary
//  impact on the sort order of ids. The sequence # is next most significant
//  as it is the tie-breaker when the time portions are equivalent.
//
//  Note that the lower 64 bits are basically random and not specifically
//  useful for ordering, although they play their part when the upper 64-bits
//  are equivalent between two ids. Again, the ordering outcome in this
//  situation is somewhat random, but generally somewhat repeatable (hardware
//  id should be consistent and stable a vast majority of the time).
//  ---------------------------------------------------------------------------

// NewOvertFlakeID53Synthesizer creates an instance of generator (which implements Generator) and
// restricts the most-significant 8 bytes to use 53 bits (float64 precision), with 41 bytes for
// time, and 12 bytes for sequence #. Epoch is always OvertoneEpochMs to maximize the range
// of the 41 bits to 127 years
//
func NewOvertFlakeID53Synthesizer(hardwareID HardwareID, processID int) OvertFlakeIDGenerator {
	// binary.BigEndian.Uint64 won't work on a []byte < len(8) so we need to
	// copy our 6-byte hardwareID into the most-signficant bits
	tempBytes := make([]byte, 8)
	copy(tempBytes[0:6], hardwareID[0:6])

	return &overtFlakeIDSynthesizer{
		epoch:        OvertoneEpochMs,
		sequenceBits: SequenceBits53,
		sequenceMask: uint64(int64(-1) ^ (int64(-1) << SequenceBits53)),
		upperMask:    MSBMask53,
		hardwareID:   hardwareID,
		processID:    processID & 0xFFFF,
		machineID:    binary.BigEndian.Uint64(tempBytes) | uint64(processID&0xFFFF),
	}
}

// NewOvertFlakeGenerator53 creates an instance of generator which implements Generator
func NewOvertFlakeGenerator53(hardwareID HardwareID, processID int, waitForTime int64) Generator {
	// binary.BigEndian.Uint64 won't work on a []byte < len(8) so we need to
	// copy our 6-byte hardwareID into the most-signficant bits
	tempBytes := make([]byte, 8)
	copy(tempBytes[0:6], hardwareID[0:6])

	return &generator{
		idGen:    NewOvertFlakeID53Synthesizer(hardwareID, processID),
		lastTime: waitForTime,
	}
}
