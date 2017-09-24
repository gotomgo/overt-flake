package flake

// Generator is the primary interface for flake ID generation
type Generator interface {
	Epoch() int64
	HardwareID() int64
	ProcessID() int64

	Generate(count int) ([]byte, error)
}
