package port_allocation_pool

import (
	"sync"
	"testing"
)

func TestRandomAllocation(t *testing.T) {
	// Fills pool
	portPool, _ := NewPortAllocationPool(1, 65535)

	for i := 0; i < portPool.maxNumberOfPorts; i++ {
		_, err := portPool.AllocateRandomPort()
		if err != nil {
			t.Errorf("Unexpected error while allocating port in poool %d.", i)

		}
	}

	// Add another port will cause an error
	_, err := portPool.AllocateRandomPort()
	if err == nil {
		t.Errorf("That port shoudl not have been alocated.")
	}

	// Removes port 21
	err = portPool.DeallocatePort(21)
	if err != nil {
		t.Errorf("Port deallocation failed when it should have been succeeded")
	}

	// Adds port 21 by random allocating only reamining spot
	port21, lastErr := portPool.AllocateRandomPort()
	if lastErr != nil {
		t.Errorf("Not expected error")
	}
	if port21 != 21 {
		t.Errorf("Port 21 expected to be returned, but %d was returned instead", port21)
	}

}

func TestManualAllocation(t *testing.T) {
	// Fills pool
	portPool, _ := NewPortAllocationPool(1, 65535)

	// Alocate port 21
	err := portPool.AllocatePort(21)
	if err != nil {
		t.Errorf("Port 21 expected to be allocated")
	}

	// Tries to allocate it again
	err = portPool.AllocatePort(21)
	if err == nil {
		t.Errorf("Port 21 was allocated twice!")
	}

	// Deallcoate
	err = portPool.DeallocatePort(21)
	if err != nil {
		t.Errorf("Failed to deallocate port 21!")
	}

	// Deallcoate twice (shoudln't generate erros)
	err = portPool.DeallocatePort(21)
	if err != nil {
		t.Errorf("Dealocating port multiple times should not be a problem!")
	}

	// Allocated port 21 again
	err = portPool.AllocatePort(21)
	if err != nil {
		t.Errorf("Port 21 expected to be allocated!")
	}

	// Mix some random allocation in the baking
	for i := 0; i < (portPool.maxNumberOfPorts - 1); i++ {
		_, err := portPool.AllocateRandomPort()
		if err != nil {
			t.Errorf("Unexpected error while allocating port in poool %d.", i)

		}
	}

	// Add another port will cause an error
	_, err = portPool.AllocateRandomPort()
	if err == nil {
		t.Errorf("That port shoudl not have been alocated.")
	}

}

func TestConcurrency(t *testing.T) {
	// Fills pool
	portPool, _ := NewPortAllocationPool(1, 65535)
	var wg sync.WaitGroup

	for i := 0; i < portPool.maxNumberOfPorts; i++ {
		wg.Add(1)
		go allocatePort(portPool, &wg, t, i)
	}

	wg.Wait()

	// New allocation shoud cauisa an error
	_, err := portPool.AllocateRandomPort()
	if err == nil {
		t.Errorf("That port shoudl not have been alocated.")
	}
}

func allocatePort(portPool *PortAllocationPool, wg *sync.WaitGroup, t *testing.T, i int) {
	_, err := portPool.AllocateRandomPort()
	if err != nil {
		t.Errorf("Unexpected error while allocating port in poool %d.", i)

	}
	wg.Done()
}
