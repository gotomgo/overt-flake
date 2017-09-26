package flake

type fixedHardwareIDProvider struct {
	hardwareID HardwareID
}

// NewFixedHardwareIDProvider creates an instance of fixedHardwareIDProvider
func NewFixedHardwareIDProvider(hardwareID HardwareID) HardwareIDProvider {
	return &fixedHardwareIDProvider{
		hardwareID: hardwareID,
	}
}

func (fixed *fixedHardwareIDProvider) GetHardwareID(byteSize int) ([]byte, error) {
	if byteSize > len(fixed.hardwareID) {
		return nil, ErrInvalidSizeForHardwareAddress
	}

	return fixed.hardwareID, nil
}
