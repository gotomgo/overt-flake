package flake

import (
	"encoding/binary"
	"os"
)

const (
	// DefaultSequenceBits is the default # of bits used for sequence #'s
	DefaultSequenceBits uint64 = 16
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
//  | 0 | 1 | 2 | 3 | 4 | 5 |  6  |  7  |  8  | 9 | A | B | C | D |  E  |  F  |
//  ---------------------------------------------------------------------------
//  |           48 bits     |  16 bits  |         48 bits         |  16 bits  |
//  ---------------------------------------------------------------------------
//  |          timestamp    |  sequence |        HardwareID       | ProcessID |
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

type overtFlakeIDSynthesizer struct {
	epoch        int64
	sequenceBits uint64
	sequenceMask uint64
	hardwareID   HardwareID
	processID    int
	machineID    uint64
}

// NewOvertFlakeIDSynthesizer creates an instance of generator (which implements Generator) and
// allows the # of sequence bits to be specified (16 is standard)
//
// Notes
//
// Setting a value of sequenceBits > 22 will result in unacceptable time truncation
func NewOvertFlakeIDSynthesizer(epoch int64, sequenceBits uint64, hardwareID HardwareID, processID int) OvertFlakeIDGenerator {
	// binary.BigEndian.Uint64 won't work on a []byte < len(8) so we need to
	// copy our 6-byte hardwareID into the most-signficant bits
	tempBytes := make([]byte, 8)
	copy(tempBytes[0:6], hardwareID[0:6])

	return &overtFlakeIDSynthesizer{
		epoch:        epoch,
		sequenceBits: sequenceBits,
		sequenceMask: uint64(int64(-1) ^ (int64(-1) << sequenceBits)),
		hardwareID:   hardwareID,
		processID:    processID & 0xFFFF,
		machineID:    binary.BigEndian.Uint64(tempBytes) | uint64(processID&0xFFFF),
	}
}

// NewOvertFlakeGeneratorWithBits creates an instance of generator (which implements Generator.) this
// "constructor" allows the # of sequence bits to be specified
//
// Notes
//
// Setting a value of seqBits > 22 will result in unacceptable time truncation
func NewOvertFlakeGeneratorWithBits(epoch int64, hardwareID HardwareID, processID int, waitForTime int64, seqBits uint64) Generator {
	// binary.BigEndian.Uint64 won't work on a []byte < len(8) so we need to
	// copy our 6-byte hardwareID into the most-signficant bits
	tempBytes := make([]byte, 8)
	copy(tempBytes[0:6], hardwareID[0:6])

	return &generator{
		idGen:    NewOvertFlakeIDSynthesizer(epoch, seqBits, hardwareID, processID),
		lastTime: waitForTime,
	}
}

// NewOvertFlakeGenerator creates an instance of generator which implements Generator
func NewOvertFlakeGenerator(epoch int64, hardwareID HardwareID, processID int, waitForTime int64) Generator {
	return NewOvertFlakeGeneratorWithBits(epoch, hardwareID, processID, waitForTime, DefaultSequenceBits)
}

// NewOvertoneEpochGenerator creates an instance of generator using the Overtone Epoch
func NewOvertoneEpochGenerator(hardwareID HardwareID) Generator {
	return NewOvertFlakeGeneratorWithBits(OvertoneEpochMs, hardwareID, os.Getpid(), 0, DefaultSequenceBits)
}

func (ofid *overtFlakeIDSynthesizer) HardwareID() HardwareID {
	return ofid.hardwareID
}

func (ofid *overtFlakeIDSynthesizer) ProcessID() int {
	return ofid.processID
}

func (ofid *overtFlakeIDSynthesizer) IDSize() int {
	return OvertFlakeIDLength
}

func (ofid *overtFlakeIDSynthesizer) SequenceBitCount() uint64 {
	return ofid.sequenceBits
}

func (ofid *overtFlakeIDSynthesizer) SequenceBitMask() uint64 {
	return ofid.sequenceMask
}

func (ofid *overtFlakeIDSynthesizer) MaxSequenceNumber() uint64 {
	return ofid.sequenceMask
}

func (ofid *overtFlakeIDSynthesizer) Epoch() int64 {
	return ofid.epoch
}

func (ofid *overtFlakeIDSynthesizer) SynthesizeID(buffer []byte, index int, time int64, sequence uint64) int {
	// time is Unix Epoch (note that this is inefficient in that delta has to be calculated for
	// each id, when the generator could do the calculation + shift 1 time per allocate)
	delta := uint64(time - ofid.epoch)

	// upper 32 are time | sequence
	var upper = (delta << ofid.sequenceBits) | (sequence & ofid.sequenceMask)

	// Write the id
	binary.BigEndian.PutUint64(buffer[index:index+8], upper)
	binary.BigEndian.PutUint64(buffer[index+8:index+16], ofid.machineID)

	// return the length of the id
	return OvertFlakeIDLength
}
