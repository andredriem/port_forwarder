package port_allocation_pool

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/golang-collections/go-datastructures/bitarray"
)

var packageRand = rand.New(rand.NewSource(0))

func init() {
	packageRand.Seed(time.Now().UTC().UnixNano())
}

type ErrorCause int

const (
	NEGATIVE_PORT_RANGE ErrorCause = iota
	POOL_START_OUT_OF_RANGE
	POOL_END_OUT_OF_RANGE
	NO_PORT_AVAILABLE
	PORT_NOT_AVAILABLE
	PORT_OUT_OF_RANGE
)

type PortAllocationPoolError struct {
	Cause ErrorCause
}

func (e *PortAllocationPoolError) Error() string {
	switch e.Cause {
	case NEGATIVE_PORT_RANGE:
		return fmt.Sprintf("PoolEnd must be bigger than PoolStart")
	case POOL_START_OUT_OF_RANGE:
		return fmt.Sprintf("PoolStart must be between 1 and 65535")
	case POOL_END_OUT_OF_RANGE:
		return fmt.Sprintf("PoolEnd must be between 1 and 65535")
	case NO_PORT_AVAILABLE:
		return fmt.Sprintf("Thera are no ports available")
	case PORT_NOT_AVAILABLE:
		return fmt.Sprintf("The requestes port is not available")
	case PORT_OUT_OF_RANGE:
		return fmt.Sprintf("The reuqested port is out of the configured range")
	default:
		return fmt.Sprintf("Unable to determine the error cause")
	}
}

type PortAllocationPool struct {
	alloactedPortsBitArray bitarray.BitArray
	allocationPoolMutex    sync.RWMutex
	forwardPortPoolStart   int
	forwardPortPoolEnd     int
	maxNumberOfPorts       int
}

func NewPortAllocationPool(poolStart int, poolEnd int) (*PortAllocationPool, error) {

	// Inclusive range between poolEnd and poolStart
	poolRange := poolEnd - poolStart + 1

	if poolRange <= 0 {
		return nil, &PortAllocationPoolError{Cause: NEGATIVE_PORT_RANGE}
	}

	if !(1 <= poolStart || poolStart <= 65535) {
		return nil, &PortAllocationPoolError{Cause: POOL_START_OUT_OF_RANGE}
	}

	if !(1 <= poolEnd || poolEnd <= 65535) {
		return nil, &PortAllocationPoolError{Cause: POOL_END_OUT_OF_RANGE}
	}

	pool := &PortAllocationPool{
		forwardPortPoolStart:   poolStart,
		forwardPortPoolEnd:     poolEnd,
		maxNumberOfPorts:       poolRange,
		alloactedPortsBitArray: bitarray.NewBitArray(uint64(poolRange)),
	}

	return pool, nil

}

func (pap *PortAllocationPool) MaxNumberOfPorts() int {
	return pap.maxNumberOfPorts
}

func (pap *PortAllocationPool) ForwardPortPoolStart() int {
	return pap.forwardPortPoolStart
}

func (pap *PortAllocationPool) ForwardPortPoolEnd() int {
	return pap.forwardPortPoolEnd
}

func (pap *PortAllocationPool) AllocateRandomPort() (int, error) {
	pap.allocationPoolMutex.Lock()
	defer pap.allocationPoolMutex.Unlock()

	// Tries to generate a pseudo-random positive number
	generatedPortPoolnumber := uint64(packageRand.Intn(pap.maxNumberOfPorts))
	generatedPort := pap.forwardPortPoolStart + int(generatedPortPoolnumber)

	poolStatus, _ := pap.alloactedPortsBitArray.GetBit(generatedPortPoolnumber)
	if poolStatus == false {
		pap.alloactedPortsBitArray.SetBit(generatedPortPoolnumber)
		return generatedPort, nil
	}

	//Bad scenario, now we will loop trought the bool until we find a good spot
	maxNumberOfPorts := uint64(pap.maxNumberOfPorts)
	portPoolCandidate := (generatedPortPoolnumber + 1) % maxNumberOfPorts
	for portPoolCandidate != generatedPortPoolnumber {

		poolStatus, _ := pap.alloactedPortsBitArray.GetBit(portPoolCandidate)
		if poolStatus == false {
			pap.alloactedPortsBitArray.SetBit(portPoolCandidate)
			return pap.forwardPortPoolStart + int(portPoolCandidate), nil
		}

		portPoolCandidate = (portPoolCandidate + 1) % maxNumberOfPorts
	}

	return 0, &PortAllocationPoolError{NO_PORT_AVAILABLE}
}

func (pap *PortAllocationPool) AllocatePort(port int) error {
	pap.allocationPoolMutex.Lock()
	defer pap.allocationPoolMutex.Unlock()

	portPoolNumber := uint64(port - pap.forwardPortPoolStart)
	poolStatus, err := pap.alloactedPortsBitArray.GetBit(uint64(portPoolNumber))

	if err != nil {
		return &PortAllocationPoolError{PORT_OUT_OF_RANGE}
	}

	if poolStatus == false {
		pap.alloactedPortsBitArray.SetBit(portPoolNumber)
		return nil
	}

	return &PortAllocationPoolError{PORT_NOT_AVAILABLE}

}

func (pap *PortAllocationPool) DeallocatePort(port int) error {
	pap.allocationPoolMutex.Lock()
	defer pap.allocationPoolMutex.Unlock()

	err := pap.alloactedPortsBitArray.ClearBit(uint64(port - pap.forwardPortPoolStart))

	if err != nil {
		return &PortAllocationPoolError{PORT_OUT_OF_RANGE}
	}

	return nil
}
