package flake

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOvertFlakeID(t *testing.T) {
	gen := NewOvertoneEpochGenerator(testHardwareID)
	idBytes, err := gen.Generate(1)
	assert.NoError(t, err)

	id := NewOvertFlakeID(idBytes)

	assert.Equal(t, idBytes, id.Bytes())

	upper := binary.BigEndian.Uint64(idBytes[0:8])
	lower := binary.BigEndian.Uint64(idBytes[8:16])

	assert.Equal(t, upper, id.Upper())
	assert.Equal(t, lower, id.Lower())

	assert.Equal(t, upper>>16, id.Timestamp())
	assert.Equal(t, uint16(upper&0xFFFF), id.Interval())

	assert.Equal(t, HardwareID(idBytes[8:14]), id.HardwareID())
	assert.Equal(t, uint16(lower&0xFFFF), id.ProcessID())
	assert.Equal(t, lower, id.MachineID())

	bigInt := id.ToBigInt()
	assert.NotNil(t, bigInt)
	assert.Equal(t, bigInt.String(), id.String())
}
