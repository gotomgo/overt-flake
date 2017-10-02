package flake

import (
	"sync"
	"time"
)

// generator is an implementtion of Generator and acts as a host for an
// IDGenerator which creates and writes the identifiers for the Generator,
// while the Generator worries about logic around time intervals and allocating
// sequence #'s', and providing APIs for generating identifiers in bulk.
//
// For convienence, all implementations of Generator also implement IDGenerator,
// and for generator, all IDGenerator methods are proxied to the actual
// implementation of IDGenerator
//
//	- idGen is the IDGenerator that forms and writes the ID on the generator's
//		behalf
//	lastTime is the time interval when ids were last allocated. This value can
// 		be saved periodically to know the minimum re-start time if the server
//		crashes and needs to be restarted
//	sequence is the sequence # for the current interval. It resets each
//		millisecond (but only if 1 or more ids are being generated during
//		the interval)
type generator struct {
	idGen IDGenerator

	lastTime int64
	sequence uint64

	mutex sync.Mutex
}

// IDSize implements IDGenerator.IDSize() and is a proxy to the underlying
// IDGenerator
func (gen *generator) IDSize() int {
	return gen.idGen.IDSize()
}

// SequenceBitCount implements IDGenerator.SequenceBitCount() and is a proxy to
// the underlying IDGenerator
func (gen *generator) SequenceBitCount() uint64 {
	return gen.idGen.SequenceBitCount()
}

// SequenceBitMask implements IDGenerator.SequenceBitMask() and is a proxy to
// the underlying IDGenerator
func (gen *generator) SequenceBitMask() uint64 {
	return gen.idGen.SequenceBitMask()
}

// MaxSequenceNumber implements IDGenerator.MaxSequenceNumber() and is a proxy to
// the underlying IDGenerator
func (gen *generator) MaxSequenceNumber() uint64 {
	return gen.idGen.MaxSequenceNumber()
}

// IDGenerator is the IDGenerator used by the generator
func (gen *generator) IDGenerator() IDGenerator {
	return gen.idGen
}

// SynthesizeID uses the IDGenerator to construct an write an id to a buffer
func (gen *generator) SynthesizeID(buffer []byte, index int, time int64, sequence uint64) int {
	return gen.idGen.SynthesizeID(buffer, index, time, sequence)
}

// Epoch returns the epoch used for the identifers created by the generator
// (via the IDGenerator)
func (gen *generator) Epoch() int64 {
	return gen.idGen.Epoch()
}

// LastAllocatedTime is the last Unix Epoch value that one or more ids
// are known to have been generated
func (gen *generator) LastAllocatedTime() int64 {
	gen.mutex.Lock()
	defer gen.mutex.Unlock()

	return gen.lastTime
}

func (gen *generator) GenerateAsStream(count int, buffer []byte, callback func(int, []byte) error) (totalAllocated int, err error) {
	if len(buffer) < gen.IDSize() {
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
			return
		}

		// for each ID that was allocated, write the bytes for the ID to
		// the results array
		for j := uint64(0); j < allocated; j++ {
			index += gen.SynthesizeID(buffer, index, interval, j)

			// buffer is full
			if index >= len(buffer) {
				err = callback(index/gen.IDSize(), buffer)
				if err != nil {
					return
				}

				// more were delivered so update our return value
				totalAllocated += int(index / gen.IDSize())

				// back to beginning of the buffer
				index = 0
			}
		}

		// partial buffer fill
		if index > 0 {
			err = callback(index/gen.IDSize(), buffer)
			if err != nil {
				return
			}

			// more were delivered so update our return value
			totalAllocated += int(index / gen.IDSize())

			// back to beginning of the buffer
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
	results = make([]byte, gen.IDSize()*count)

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
	if uint64(count) > gen.MaxSequenceNumber() {
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

	gen.lastTime = current

	// allocated the request # of items, or whatever is remaining for this cycle
	var allocated uint64
	if uint64(count) > gen.sequence-gen.MaxSequenceNumber() {
		allocated = gen.sequence - gen.MaxSequenceNumber()
	} else {
		allocated = uint64(count)
	}

	// advance the sequence for the # of items allocated
	gen.sequence = (gen.sequence + allocated) & gen.SequenceBitMask()

	return allocated, current, nil
}

// timestamp returns the # of milliseconds that have passed since
// the unix epoch
func timestamp() int64 {
	return time.Now().UnixNano() / 1e6
}
