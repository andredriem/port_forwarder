package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"

	"./redirect"

	"github.com/gorilla/mux"
)

var ipTablesExecutable (string)
var allowPermanentRules bool
var exposedPortsStartRange int
var exposedPortsEndRange int

func loadEnviromentVariables() {

}

// Main function
func main() {
	// Init router
	localIpTablesExecutable, err := exec.LookPath("iptables")
	if err != nil {
		panic(err)
	}

	ipTablesExecutable = localIpTablesExecutable

	r := mux.NewRouter()

	// Route handles & endpoints
	r.HandleFunc("/create_port_forward", createPortForward).Methods("POST")

	// Start server
	log.Fatal(http.ListenAndServe(":3000", r))
}

func createPortForward(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	aRedirect, _ := redirect.NewRedirectFromJson(r.Body)

	redirect.AddRedirectToFirewall(aRedirect)

	json.NewEncoder(w).Encode("")
}
