package forwarded_port_pool

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

var forwardPortPoolStart int
var forwardPortPoolEnd int
var MaxNumberOfPorts int
var packageRand = rand.New(rand.NewSource(0))

type AllocationPortArray []bool

var GlobalAllocationPortArray AllocationPortArray
var allocationPoolMutex sync.RWMutex

func (allocationPortArray AllocationPortArray) allocateRandomPort() (int, error) {

	// Tries to generate a pseudo-random positive number
	generatedPortPoolnumber := packageRand.Intn(MaxNumberOfPorts)
	generatedPort := forwardPortPoolStart + generatedPortPoolnumber
	//locks mutex for writing
	allocationPoolMutex.Lock()
	defer allocationPoolMutex.Unlock()

	if GlobalAllocationPortArray[generatedPortPoolnumber] == false {
		GlobalAllocationPortArray[generatedPortPoolnumber] = true
		return generatedPort, nil
	}

	//Bad scenario, now we will loop trought the bool until we find a good spot
	portPoolCandidate := (generatedPortPoolnumber + 1) % MaxNumberOfPorts
	portCandidate := portPoolCandidate + forwardPortPoolStart
	for portCandidate != generatedPortPoolnumber {
		if GlobalAllocationPortArray[portPoolCandidate] == false {
			GlobalAllocationPortArray[portPoolCandidate] = true
			return portCandidate, nil
		}

		portPoolCandidate = (portPoolCandidate + 1) % MaxNumberOfPorts
		portCandidate = portPoolCandidate + forwardPortPoolStart
	}

	return 0, errors.New("No Ports available")
}

func (allocationPortArray AllocationPortArray) allocatePort(port int) error {
	if !isPortInRange(port) {
		return errors.New("Port is not in range!")
	}

	portPoolNumber := port - forwardPortPoolStart

	//Tries to allocate port
	allocationPoolMutex.Lock()
	defer allocationPoolMutex.Unlock()
	if GlobalAllocationPortArray[portPoolNumber] == false {
		GlobalAllocationPortArray[portPoolNumber] = true
		return nil
	}

	return errors.New("Port not available!")

}

func (allocationPortArray AllocationPortArray) deallocatePort(port int) error {
	if !isPortInRange(port) {
		return errors.New("Port is not in range!")
	}

	portPoolNumber := port - forwardPortPoolStart
	allocationPoolMutex.Lock()
	defer allocationPoolMutex.Unlock()
	GlobalAllocationPortArray[portPoolNumber] = false
	return nil
}

func isPortInRange(port int) bool {
	return forwardPortPoolStart <= port || port <= forwardPortPoolEnd
}

func init() {
	//Remove this when I find a way to run init after testing
	if true {
		os.Setenv("FORWARD_PORT_POOL_START", "20")
		os.Setenv("FORWARD_PORT_POOL_END", "2000")
	}

	assingPoolFromEnviroment("FORWARD_PORT_POOL_START", &forwardPortPoolStart)
	assingPoolFromEnviroment("FORWARD_PORT_POOL_END", &forwardPortPoolEnd)

	if forwardPortPoolStart >= forwardPortPoolEnd {
		panic(
			fmt.Sprintf(
				"FORWARD_PORT_POOL_START(%d) MUST BE AT LEAST A NUMBER BELOW FORWARD_PORT_END(%d).",
				forwardPortPoolStart,
				forwardPortPoolEnd,
			),
		)
	}

	MaxNumberOfPorts = (forwardPortPoolEnd - forwardPortPoolStart + 1)
	GlobalAllocationPortArray = make(AllocationPortArray, MaxNumberOfPorts)

	// Seed local rand
	packageRand.Seed(time.Now().UTC().UnixNano())

	//Init mutexes

}

func assingPoolFromEnviroment(envVariableName string, poolVariable *int) {
	var err error
	var convertedPoolNumber int
	envVariable := os.Getenv(envVariableName)

	if envVariable == "" {
		panic(fmt.Sprintf("Error: Required enviroment variable %s is missing!", envVariableName))
	}

	convertedPoolNumber, err = strconv.Atoi(envVariable)
	if err != nil || convertedPoolNumber <= 0 || convertedPoolNumber > 65535 {
		panic(fmt.Sprintf("Error: %s must be a valid VLAN between 0 and 65535", envVariableName))
	}

	*poolVariable = convertedPoolNumber
}
