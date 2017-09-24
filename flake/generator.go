package flake

import (
	"errors"
	"sync"
)

type generator struct {
	hardwareID int64
	processID  int64
	epoch      int64
	lastTime   int64
	sequence   int64
	mutex      sync.Mutex
}

// NewGenerator creates an instance of generator which implements Generator
func NewGenerator(epoch, hardwardID, processID int64) Generator {
	return &generator{
		epoch:      epoch,
		hardwareID: hardwardID,
		processID:  processID,
	}
}

// NewOvertoneGenerator creates an instance of generator using the Overtone Epoch
func NewOvertoneGenerator(hardwardID, processID int64) Generator {
	return NewGenerator(OvertoneEpochMs, hardwardID, processID)
}

func (gen *generator) Epoch() int64 {
	return gen.epoch
}

func (gen *generator) HardwareID() int64 {
	return gen.hardwareID
}

func (gen *generator) ProcessID() int64 {
	return gen.processID
}

func (gen *generator) Generate(count int) ([]byte, error) {
	return nil, errors.New("Not Implemented")
}
