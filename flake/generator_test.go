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
	assert.Equal(t, os.Getpid(), gen.ProcessID(), "Expecting generator.ProcessID() to == os.Getpid()")
	assert.Equal(t, testHardwareID, gen.HardwareID(), "Expecting generator.HardwareID() to == testHardwareID")

	// remember when we start the gen so we can compare the timestamp in the id for >=
	startTime := time.Now().UTC().UnixNano() / 1e6

	id, err := gen.Generate(1)
	assert.NoError(t, err)
	assert.NotNil(t, id, "Expecting id to be non-nill if err == nil")
	assert.Equal(t, OvertFlakeIDLength, len(id), "Expecting id length to == %d, not %d", OvertFlakeIDLength, len(id))

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
