package flake

import "net"

type simpleMacHardwareIDProvider struct{}

// NewSimpleMacHardwareIDProvider creates a new instance of simpleMacHardwareIDProvider
// which implements HardwareIDProvider
func NewSimpleMacHardwareIDProvider() HardwareIDProvider {
	return &simpleMacHardwareIDProvider{}
}

func (mac *simpleMacHardwareIDProvider) GetHardwareID(byteSize int) ([]byte, error) {
	// On the lower bound we presume the caller(s) will do something reasonable. For the
	// upper bound we are limited to the 6 bytes comprising the MAC address
	if byteSize > 6 {
		return nil, ErrInvalidSizeForHardwareAddress
	}

	inets, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	if len(inets) == 0 {
		return nil, ErrNoNetworkInterfaces
	}

	for _, net := range inets {
		if net.HardwareAddr != nil {
			return net.HardwareAddr[0:byteSize], nil
		}
	}

	return nil, ErrNoHardwareAddresses
}
