package flake

import (
	"encoding/binary"
	"os"
	"sync"
	"time"
)

//  ---------------------------------------------------------------------------
//  Layout - Big Endian
//  ---------------------------------------------------------------------------
//  [0:6]   48 bits | Upper 48 bits of timestamp (milliseconds since the epoch)
//  [6:8]   16 bits | a per-interval sequence # (interval == 1 millisecond)
//  [9:14]  48 bits | a hardware id
//  [14:16] 16 bits | process ID
//  ---------------------------------------------------------------------------
//  | 0 | 1 | 2 | 3 | 4 | 5 | 6 |  7  |  8  | 9 | A | B | C | D |  E  |  F  |
//  ---------------------------------------------------------------------------
//  |           48 bits         |  16 bits  |     48 bits       |  16 bits  |
//  ---------------------------------------------------------------------------
//  |          timestamp        |  interval |    HardwareID     | ProcessID |
//  ---------------------------------------------------------------------------
//  Notes
//  ---------------------------------------------------------------------------
//  The time bits are the most signficant bits because they have the primary
//  impact on the sort order of ids. The interval/seq # is next most significant
//  as it is the tie-breaker when the time portions are equivalent.
//
//  Note that the lower 64 bits are basically random and not specifically
//  useful for ordering, although they play their party when the upper 64-bits
//  are equivalent between two ideas. Again, the ordering outcome in this
//  situation is somewhat random, but generally somewhat repeatable (hardware
//  id should be consistent and stable a vast majority of the time).
//  ---------------------------------------------------------------------------

var sequenceBits uint64 = 16
var sequenceMask = uint64(int64(-1) ^ (int64(-1) << sequenceBits))
var maxSequenceNumber = uint64(^sequenceMask)

type generator struct {
	epoch      int64
	hardwareID HardwareID
	processID  int
	lastTime   int64
	sequence   uint64
	mutex      sync.Mutex

	machineID uint64
}

// NewGenerator creates an instance of generator which implements Generator
func NewGenerator(epoch int64, hardwareID HardwareID, processID int) Generator {
	// binary.BigEndian.Uint64 won't work on a []byte < len(8) so we need to
	// copy our 6-byte hardwareID into the most-signficant bits
	tempBytes := make([]byte, 8)
	copy(tempBytes[0:6], hardwareID[0:6])

	return &generator{
		epoch:      epoch,
		hardwareID: hardwareID,
		processID:  processID & 0xFFFF,
		machineID:  binary.BigEndian.Uint64(tempBytes) | uint64(processID&0xFFFF),
	}
}

// NewOvertoneEpochGenerator creates an instance of generator using the Overtone Epoch
func NewOvertoneEpochGenerator(hardwareID HardwareID) Generator {
	return NewGenerator(OvertoneEpochMs, hardwareID, os.Getpid())
}

func (gen *generator) Epoch() int64 {
	return gen.epoch
}

func (gen *generator) HardwareID() HardwareID {
	return gen.hardwareID
}

func (gen *generator) ProcessID() int {
	return gen.processID
}

func (gen *generator) GenerateAsStream(count int, buffer []byte, callback func(int, []byte) error) (totalAllocated int, err error) {
	if len(buffer) < OvertFlakeIDLength {
		return 0, ErrBufferTooSmall
	}

	// while we still have ids to allocate/generate
	for count > 0 {
		var allocated uint64
		var interval int64
		var index int

		// allocate as many ids as available up to count
		allocated, interval, err = gen.allocate(count)
		if err != nil {
			return totalAllocated, err
		}

		totalAllocated += int(allocated)

		// calculate the delta between the interval (Unix Epoch in milliseconds)
		// and the epoch being used for id generation
		delta := uint64((interval - gen.epoch) << 16)

		// for each ID that was allocated, write the bytes for the ID to
		// the results array
		for j := uint64(0); j < allocated; j++ {
			var upper = delta | (j & sequenceMask)
			binary.BigEndian.PutUint64(buffer[index:index+8], upper)
			binary.BigEndian.PutUint64(buffer[index+8:index+16], gen.machineID)
			index += 16

			// buffer is full
			if index >= len(buffer) {
				err = callback(index/16, buffer)
				if err != nil {
					return
				}
				index = 0
			}
		}

		// partial buffer fill
		if index > 0 {
			callback(index/16, buffer)
			if err != nil {
				return
			}
			index = 0
		}

		count -= int(allocated)
	}

	return
}

// Generate uses allocate to allocate as many ids as required, and writes
// each id into a contiguous []byte
func (gen *generator) Generate(count int) (results []byte, err error) {
	// allocate a buffer that will hold count IDs
	results = make([]byte, OvertFlakeIDLength*count)

	var allocated int

	// use the stream API but because our buffer can hold all allocated ids we dont need
	// to react in the callback
	allocated, err = gen.GenerateAsStream(count, results, func(allocated int, ids []byte) error {
		return nil
	})

	// we do not want to return a partial result
	if (allocated != count) || (err != nil) {
		results = nil
	}

	return
}

// allocate does all the magic of time and sequence management. It does not
// perfomm the generation of the ids, but provides the data required to do so
func (gen *generator) allocate(count int) (uint64, int64, error) {
	if uint64(count) > maxSequenceNumber {
		return 0, 0, ErrTooManyRequested
	}

	// We need to take the lock so we can manipulate the generator state
	gen.mutex.Lock()
	defer gen.mutex.Unlock()

	// current time since Unix Epoch in milliseconds
	current := timestamp()

	// Is time going backwards? Thats a problem
	if current < gen.lastTime {
		return 0, 0, ErrTimeIsMovingBackwards
	}

	if gen.lastTime != current {
		gen.lastTime = current
		gen.sequence = 0
	} else {
		// When all the ids have been allocated for this interval then we end up
		// here and we need to spin for the next cycle
		if gen.sequence == 0 {
			for current <= gen.lastTime {
				current = timestamp()
			}
		}
	}

	// allocated the request # of items, or whatever is remaining for this cycle
	var allocated uint64
	if uint64(count) > gen.sequence-sequenceMask {
		allocated = gen.sequence - sequenceMask
	} else {
		allocated = uint64(count)
	}

	// advance the sequence for the # of items allocated
	gen.sequence = (gen.sequence + allocated) & sequenceMask

	return allocated, current, nil
}

// timestamp returns the # of milliseconds that have passed since
// the unix epoch
func timestamp() int64 {
	return time.Now().UnixNano() / 1e6
}
