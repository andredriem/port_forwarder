package forwarded_port_pool

import (
	"testing"
)

func TestRandomAllocation(t *testing.T) {
	// Fills pool
	for i := 0; i < MaxNumberOfPorts; i++ {
		_, err := GlobalAllocationPortArray.allocateRandomPort()
		if err != nil {
			t.Errorf("Unexpected error while allocating port in poool %d.", i)

		}
	}

	// Add another port will cause an error
	_, err := GlobalAllocationPortArray.allocateRandomPort()
	if err == nil {
		t.Errorf("That port shoudl not have been alocated.")
	}

	// Removes port 21
	err = GlobalAllocationPortArray.deallocatePort(21)
	if err != nil {
		t.Errorf("Port deallocation failed when it should have been succeeded")
	}

	// Adds port 21 by random allocating only reamining spot
	port21, lastErr := GlobalAllocationPortArray.allocateRandomPort()
	if lastErr != nil {
		t.Errorf("Not expected error")
	}
	if port21 != 21 {
		t.Errorf("Port 21 expected to be returned, but %d was returned instead", port21)
	}

}
