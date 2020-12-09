package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func executeAllocateRandomPortRequest(ip string, destPort int, ttl int) *httptest.ResponseRecorder {
	r, _ := http.NewRequest("POST", "allocate_random_port/", strings.NewReader(
		fmt.Sprintf(
			"{\"destIp\":\"%s\",\"destPort\":%d,\"ttlInSeconds\":%d}",
			ip,
			destPort,
			ttl,
		),
	))
	w := httptest.NewRecorder()

	allocateRandomPort(w, r)
	return w
}

func TestAllocateRandomPort(t *testing.T) {

	portRangeStart := 10
	portRangeEnd := 19
	poolRange := portRangeEnd - portRangeStart + 1

	os.Setenv("ALLOW_PERMANENT_RULES", "false")
	os.Setenv("EXPOSED_PORT_START_RANGE", strconv.Itoa(portRangeStart))
	os.Setenv("EXPOSED_PORT_END_RANGE", strconv.Itoa(portRangeEnd))
	prepareApplication()

	//test invalid ip
	w := executeAllocateRandomPortRequest("192.66.3.999", 20, 20000)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Request expected to return an error")
	}

	//test invalid destPort
	w = executeAllocateRandomPortRequest("192.66.3.1", 9000000, 20000)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Request expected to return an error")
	}

	//test invalid ttl
	w = executeAllocateRandomPortRequest("192.66.3.1", 20, -32)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Request expected to return an error")
	}

	//Ocuppy all ports
	for i := 0; i < poolRange; i++ {
		w = executeAllocateRandomPortRequest("192.66.3.1", 20, 2)
		if w.Code != http.StatusOK {
			t.Errorf("Unexpected error on iteration %d.", i)

		}
	}

	//Next port should generate an error
	w = executeAllocateRandomPortRequest("192.66.3.1", 20, 2)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Request expected to return an error")
	}

	time.Sleep(3 * time.Second)
	//Allocation now should be successfull
	w = executeAllocateRandomPortRequest("192.66.3.1", 20, 2)
	if w.Code != http.StatusOK {
		t.Errorf("Unexpected error")
	}

	// Wait for rules to clear
	time.Sleep(3 * time.Second)

}
