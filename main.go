package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"./port_allocation_pool"
	"./redirect"

	"github.com/gorilla/mux"
)

var ipTablesExecutable (string)
var allowPermanentRules bool
var exposedPortsStartRange int
var exposedPortsEndRange int
var exposedPortAllocationPool *port_allocation_pool.PortAllocationPool

func loadEnviromentVariables() {
	// Ensure all enviroment vairables are loaded correctly
	var err error
	var tempExposedPortRange int64

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

// Main function
func main() {
	var err error
	loadEnviromentVariables()

	exposedPortAllocationPool, err = port_allocation_pool.NewPortAllocationPool(exposedPortsStartRange, exposedPortsEndRange)
	if err != nil {
		switch err.(*port_allocation_pool.PortAllocationPoolError).Cause {
		case port_allocation_pool.NEGATIVE_PORT_RANGE:
			panic("EXPOSED_PORT_END_RANGE mus be bigger than EXPOSED_PORT_START_RANGE")
		case port_allocation_pool.POOL_START_OUT_OF_RANGE:
			panic("EXPOSED_PORT_START_RANGE mus be a number between 1 and 65535")
		case port_allocation_pool.POOL_END_OUT_OF_RANGE:
			panic("EXPOSED_PORT_END_RANGE mus be a number between 1 and 65535")
		default:
			panic("UNKOWN ERROR")
		}
	}

	r := mux.NewRouter()

	// Route handles & endpoints
	r.HandleFunc("/create_port_forward", createPortForward).Methods("POST")

	// Start server
	log.Fatal(http.ListenAndServe(":3000", r))
}

func createPortForward(w http.ResponseWriter, r *http.Request) {
	var err error
	w.Header().Set("Content-Type", "application/json")

	// Parse redirect object
	aRedirect, err := redirect.NewRedirectFromJson(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("")
		return
	}

	// Register redirect port in pool
	err = exposedPortAllocationPool.AllocatePort(aRedirect.ForwardedPort)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("")
		return
	}

	// Effective register port in iptables
	err = aRedirect.AddRedirectToFirewall()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("")
		// Reopens port to allocation since rule fails
		exposedPortAllocationPool.DeallocatePort(aRedirect.ForwardedPort)
		return
	}

	json.NewEncoder(w).Encode("")
}
