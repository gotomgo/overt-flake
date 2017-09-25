package flake

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleMAC(t *testing.T) {
	hardwareIDProvider := NewSimpleMacHardwareIDProvider()
	assert.NotNil(t, hardwareIDProvider)

	hardwareID, err := hardwareIDProvider.GetHardwareID(MACAddressLength)
	if err != nil {
		// maybe a legit error occurred (no interaces, not hardware addresses)
		if err != ErrNoHardwareAddresses && err != ErrNoNetworkInterfaces {
			assert.Fail(t, "Unexpected error generating hardware id", "Unexpected error occured: %s", err)
		}

		return
	}

	assert.Equal(t, MACAddressLength, len(hardwareID), "Expecting length of hardware ID to be %d, not %d", MACAddressLength, len(hardwareID))
}

func TestMACBadRequestLength(t *testing.T) {
	hardwareIDProvider := NewSimpleMacHardwareIDProvider()
	assert.NotNil(t, hardwareIDProvider)

	_, err := hardwareIDProvider.GetHardwareID(MACAddressLength - 1)
	assert.Equal(t, ErrInvalidSizeForHardwareAddress, err, "Expecting error to be %s, not %s", ErrInvalidSizeForHardwareAddress, err)
}
