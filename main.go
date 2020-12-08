package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"port_forwarder/port_allocation_pool"
	"port_forwarder/redirect"

	"github.com/gorilla/mux"
)

var ipTablesExecutable (string)
var allowPermanentRules bool
var exposedPortsStartRange int
var exposedPortsEndRange int
var serverPort int
var exposedPortAllocationPool *port_allocation_pool.PortAllocationPool
var localIp = "127.0.0.1"

func loadEnviromentVariables() {
	// Ensure all enviroment vairables are loaded correctly
	var err error
	var tempExposedPortRange int64

	var tempServerPort int64
	tempServerPort, err = strconv.ParseInt(os.Getenv("SERVER_PORT"), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("failed to convert SERVER_PORT to integer"))
	}
	serverPort = int(tempServerPort)
	if !(1 <= serverPort || serverPort <= 65535) {
		panic("SERVER_PORT must be a number between 1 and 65535")
	}

	allowPermanentRules, err = strconv.ParseBool(os.Getenv("ALLOW_PERMANENT_RULES"))
	if err != nil {
		panic(fmt.Sprintf("failed to convert ALLOW_PERMANENT_RULES to booleand"))
	}

	tempExposedPortRange, err = strconv.ParseInt(os.Getenv("EXPOSED_PORT_START_RANGE"), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("failed to convert EXPOSED_PORT_START_RANGE to integer"))
	}
	exposedPortsStartRange = int(tempExposedPortRange)

	tempExposedPortRange, err = strconv.ParseInt(os.Getenv("EXPOSED_PORT_END_RANGE"), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("failed to convert EXPOSED_PORT_END_RANGE to integer"))
	}
	exposedPortsEndRange = int(tempExposedPortRange)
}

func initializeGlobalVariables() {
	var err error
	exposedPortAllocationPool, err = port_allocation_pool.NewPortAllocationPool(exposedPortsStartRange, exposedPortsEndRange)
	if err != nil {
		switch err.(*port_allocation_pool.PortAllocationPoolError).Cause {
		case port_allocation_pool.NEGATIVE_PORT_RANGE:
			panic("EXPOSED_PORT_END_RANGE must be bigger than EXPOSED_PORT_START_RANGE")
		case port_allocation_pool.POOL_START_OUT_OF_RANGE:
			panic("EXPOSED_PORT_START_RANGE must be a number between 1 and 65535")
		case port_allocation_pool.POOL_END_OUT_OF_RANGE:
			panic("EXPOSED_PORT_END_RANGE mus be a number between 1 and 65535")
		default:
			panic("UNKOWN ERROR")
		}
	}
}

func prepareApplication() {
	loadEnviromentVariables()
	initializeGlobalVariables()
}

// Main function
func main() {
	prepareApplication()

	r := mux.NewRouter()

	// Route handles & endpoints
	r.HandleFunc("/allocate_random_port", allocateRandomPort).Methods("POST")

	// Start server
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", serverPort), r))
}

func allocateRandomPort(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	type params struct {
		DestIp       string `json:"destIp"`
		DestPort     int    `json:"destPort"`
		TtlInSeconds int    `json:"ttlInSeconds"`
	}

	var err error
	p := params{}

	//Decode params
	err = json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("")
		return
	}

	//Allocate port
	allocatedPort, err := exposedPortAllocationPool.AllocateRandomPort()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("")
		return
	}

	//Build redirect
	aRedirect, err := redirect.NewRedirect(p.DestIp, p.DestPort, localIp, allocatedPort, p.TtlInSeconds)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("")
		exposedPortAllocationPool.DeallocatePort(allocatedPort)
		return
	}

	//Add rule
	err = aRedirect.AddRedirectToFirewall(exposedPortAllocationPool)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("")
		// Reopens port to allocation since rule fails
		exposedPortAllocationPool.DeallocatePort(aRedirect.ForwardedPort)
		return
	}

	json.NewEncoder(w).Encode(fmt.Sprintf("{\"port\": %d}", allocatedPort))
}
