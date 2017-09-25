package flake

import (
	"bytes"
	"crypto/sha1"
	"net"
)

// macHardwareIDProvider implments HardwareIDProvider and produces a SHA1 of
// all available MAC hardward addresses (concatenated) to create bytes for
// use as a HardwareID
type macHardwareIDProvider struct{}

// NewMacHardwareIDProvider creates a new instance of macHardwareIDProvider
// which implements HardwareIDProvider
func NewMacHardwareIDProvider() HardwareIDProvider {
	return &macHardwareIDProvider{}
}

func (mac *macHardwareIDProvider) GetHardwareID(byteSize int) ([]byte, error) {
	// On the lower bound we presume the caller(s) will do something reasonable. For the
	// upper bound we are limited to the 20 bytes comprising the SHA1 calculated from
	// the discovered MAC addresses
	if (byteSize < MACAddressLength) || (byteSize > 20) {
		return nil, ErrInvalidSizeForHardwareAddress
	}

	inets, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	if len(inets) == 0 {
		return nil, ErrNoNetworkInterfaces
	}

	var macs [][]byte

	for _, net := range inets {
		if net.HardwareAddr != nil {
			macs = append(macs, net.HardwareAddr)
		}
	}

	if len(macs) == 0 {
		return nil, ErrNoHardwareAddresses
	}

	sha := sha1.Sum(bytes.Join(macs, nil))

	return sha[0:byteSize], err
}
