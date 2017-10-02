package flake

import "encoding/binary"

//  ---------------------------------------------------------------------------
//  Layout - Big Endian
//  ---------------------------------------------------------------------------
//
//  [0:6]  42 bits | Upper 42 bits of timestamp (milliseconds since the epoch)
//  [6:6]   5 bits | data center id
//  [6:7]   5 bits | machine id
//  [7:8]  12 bits | sequence #
//
//  ---------------------------------------------------------------------------
//  |   0   |   1   |   2   |   3   |   4   |   5   |   6   |    7    |   8   |
//  ---------------------------------------------------------------------------
//  |           42 bits                         | 5 bits | 5 bits |  12 bits  |
//  ---------------------------------------------------------------------------
//  |          timestamp                        |  dcid  |   mid  |    seq #  |
//  ---------------------------------------------------------------------------
//  Notes
//  ---------------------------------------------------------------------------
//  The time bits are the most significant bits because they have the primary
//  impact on the sort order of ids.
//  ---------------------------------------------------------------------------

type twitterFlakeIDSynthesizer struct {
	epoch        int64
	sequenceBits uint64
	idBits       uint64
	sequenceMask uint64
	machineID    int64
	dataCenterID int64
}

// NewTwitterFlakeIDSynthesizer creates an instance of generator (which implements Generator) and
// allows the # of sequence bits to be specified (16 is standard)
func NewTwitterFlakeIDSynthesizer(machineID, dataCenterID int64) IDGenerator {
	return &twitterFlakeIDSynthesizer{
		epoch:        SnowflakeEpochMs,
		sequenceBits: 12,
		idBits:       10,
		sequenceMask: uint64(int64(-1) ^ (int64(-1) << 12)),
		machineID:    machineID,
		dataCenterID: dataCenterID & 0xFFFF,
	}
}

// NewTwitterGenerator creates an instance of generator (which implements Generator.)
func NewTwitterGenerator(machineID, dataCenterID, waitForTime int64) Generator {
	return &generator{
		idGen:    NewTwitterFlakeIDSynthesizer(machineID, dataCenterID),
		lastTime: waitForTime,
	}
}

func (ofid *twitterFlakeIDSynthesizer) MachineID() int64 {
	return ofid.machineID
}

func (ofid *twitterFlakeIDSynthesizer) DataCenterID() int64 {
	return ofid.dataCenterID
}

func (ofid *twitterFlakeIDSynthesizer) IDSize() int {
	return 8
}

func (ofid *twitterFlakeIDSynthesizer) SequenceBitCount() uint64 {
	return ofid.sequenceBits
}

func (ofid *twitterFlakeIDSynthesizer) SequenceBitMask() uint64 {
	return ofid.sequenceMask
}

func (ofid *twitterFlakeIDSynthesizer) MaxSequenceNumber() uint64 {
	return ofid.sequenceMask
}

func (ofid *twitterFlakeIDSynthesizer) Epoch() int64 {
	return ofid.epoch
}

func (ofid *twitterFlakeIDSynthesizer) SynthesizeID(buffer []byte, index int, time int64, sequence uint64) int {
	// time is Unix Epoch (note that this is inefficient in that delta has to be calculated for
	// each id, when the generator could do the calculation + shift 1 time per allocate)
	id := uint64(time-ofid.epoch)<<(ofid.sequenceBits+ofid.idBits) |
		uint64(ofid.DataCenterID()<<17) |
		uint64(ofid.MachineID()<<ofid.sequenceBits) |
		uint64(sequence&ofid.sequenceMask)

	// Write the id
	binary.BigEndian.PutUint64(buffer[index:index+8], id)

	// return the length of the id
	return ofid.IDSize()
}
