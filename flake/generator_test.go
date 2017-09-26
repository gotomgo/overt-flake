package flake

import (
	"encoding/binary"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testHardwareID HardwareID

func TestMain(m *testing.M) {
	testHardwareID = make([]byte, 8)
	binary.LittleEndian.PutUint64(testHardwareID, 0x112233445566)

	retCode := m.Run()

	os.Exit(retCode)
}

func TestGenerateID(t *testing.T) {
	// Create a generator
	gen := NewOvertoneEpochGenerator(testHardwareID)
	assert.NotNil(t, gen, "Expecting generator to be allocated")
	assert.Equal(t, OvertoneEpochMs, gen.Epoch(), "Expecting generator.Epoch() to == OvertoneEpochMS")
	//assert.Equal(t, os.Getpid(), gen.ProcessID(), "Expecting generator.ProcessID() to == os.Getpid()")
	assert.Equal(t, testHardwareID, gen.HardwareID(), "Expecting generator.HardwareID() to == testHardwareID")
	assert.Equal(t, OvertFlakeIDLength, gen.IDSize(), "Expecting generator.IDSize() to == OvertFlakeIDLength")

	// remember when we start the gen so we can compare the timestamp in the id for >=
	startTime := time.Now().UTC().UnixNano() / 1e6

	id, err := gen.Generate(1)
	assert.NoError(t, err)
	assert.NotNil(t, id, "Expecting id to be non-nill if err == nil")
	assert.Equal(t, OvertFlakeIDLength, len(id), "Expecting id length to == %d, not %d", OvertFlakeIDLength, len(id))

	assert.Condition(t, func() bool {
		return gen.LastAllocatedTime() > 0
	})

	upper := binary.BigEndian.Uint64(id[0:8])
	lower := binary.BigEndian.Uint64(id[8:16])

	assert.Equal(t,
		gen.(*generator).machineID,
		lower,
		"expecting lower 64-bits of id to == generator.machineID (%d), not %d",
		gen.(*generator).machineID,
		lower)

	// because we are generating 1, and we know that we are only requestor then
	// we also know none have been allocated so the interval should be 0
	assert.Condition(t, func() bool {
		return (upper & 0xFFFF) == 0
	})

	assert.Condition(t, func() bool {
		timestamp := upper >> 16
		beginDelta := uint64(startTime - gen.Epoch())
		return timestamp >= beginDelta
	}, "Expecting upper %d >= %d", upper>>16, startTime-gen.Epoch())
}

func TestGenerateStreamIDs(t *testing.T) {
	// Create a generator
	gen := NewGenerator(OvertoneEpochMs, testHardwareID, 42, 0)

	// Create a buffer which forces the stream to provide them 1 at a time
	buffer := make([]byte, OvertFlakeIDLength)
	var called int
	totalAllocated, err := gen.GenerateAsStream(3, buffer, func(allocated int, ids []byte) error {
		called++
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, totalAllocated, "Expecting total # of ids generated to == %d, not %d", 3, totalAllocated)
	assert.Equal(t, 3, called, "Expecting total # of callbacks to be %d, not %d", 3, called)

	// Create a buffer which forces the stream to provide them 2 at a time
	buffer = make([]byte, OvertFlakeIDLength*2)
	called = 0

	// We are requesting 3 with a buffer that can hold 2, so two callbacks are expected with
	totalAllocated, err = gen.GenerateAsStream(3, buffer, func(allocated int, ids []byte) error {
		switch called {
		case 0:
			assert.Equal(t, 2, allocated, "Expecting 1st callback to have % ids, not %d", 2, allocated)
		case 1:
			assert.Equal(t, 1, allocated, "Expecting last callback to have % ids, not %d", 1, allocated)
		}
		called++
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, totalAllocated, "Expecting total # of ids generated to == %d, not %d", 3, totalAllocated)
	assert.Equal(t, 2, called, "Expecting total # of callbacks to be %d, not %d", 3, called)

}
