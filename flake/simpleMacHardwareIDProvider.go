package flake

import "net"

// simpleMacHardwareIDProvider implements HardwareIDProvider and returns the
// 1st Hardwared MAC address it can find that has enough bytes to satisfy
// the GetHardwareID request.
//
// NOTE: Use of this hardware ID provider is NOT recommended for production use
type simpleMacHardwareIDProvider struct{}

// NewSimpleMacHardwareIDProvider creates a new instance of simpleMacHardwareIDProvider
// which implements HardwareIDProvider
func NewSimpleMacHardwareIDProvider() HardwareIDProvider {
	return &simpleMacHardwareIDProvider{}
}

func (mac *simpleMacHardwareIDProvider) GetHardwareID(byteSize int) ([]byte, error) {
	// For simplicity we required the requested size to be == to the # of bytes in a MAC address (6)
	if byteSize != MACAddressLength {
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
		if len(net.HardwareAddr) >= byteSize {
			return net.HardwareAddr[0:byteSize], nil
		}
	}

	return nil, ErrNoHardwareAddresses
}
