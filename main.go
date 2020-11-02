package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Redirect struct {
	DestIp        string `json:"destIp"`
	DestPort      int    `json:"destPort"`
	ForwaredIp    string `json:"forwaredIp"`
	ForwardedPort int    `json:"forwardedPort"`
	TtlInSeconds  int    `json:"ttlInSeconds"`
}

type RuleMode int

const (
	AddRuleMode RuleMode = iota
	RemoveRuleMode
)

var ipTablesExecutable (string)

// Main function
func main() {
	// Init router
	localIpTablesExecutable, err := exec.LookPath("iptables")
	if err != nil {
		panic(err)
	}

	ipTablesExecutable = localIpTablesExecutable

	addPort("192.168.15.25", 9998, "187.45.96.132", 443, 10)

	r := mux.NewRouter()

	// Route handles & endpoints
	r.HandleFunc("/create_port_forward", createPortForward).Methods("POST")

	// Start server
	log.Fatal(http.ListenAndServe(":3000", r))
}

func createPortForward(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var redirect Redirect
	_ = json.NewDecoder(r.Body).Decode(&redirect)
	addPort(redirect.DestIp, redirect.DestPort, redirect.ForwaredIp, redirect.ForwardedPort, redirect.TtlInSeconds)

	json.NewEncoder(w).Encode("")
}

func addPort(destIp string, destPort int, forwaredIp string, forwardedPort int, ttlInSeconds int) error {
	addPortCommand := exec.Cmd{
		Path:   ipTablesExecutable,
		Args:   iptablesArguments(AddRuleMode, destIp, destPort, forwaredIp, forwardedPort),
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}

	err := addPortCommand.Run()
	fmt.Println(addPortCommand)

	if err != nil {
		log.Print(err)
	}

	duration, _ := time.ParseDuration(strconv.Itoa(ttlInSeconds) + "s")

	time.AfterFunc(duration, func() {
		removePort(destIp, destPort, forwaredIp, forwardedPort)
	})

	return err
}

func removePort(destIp string, destPort int, forwaredIp string, forwardedPort int) error {
	removePortCommand := exec.Cmd{
		Path:   ipTablesExecutable,
		Args:   iptablesArguments(RemoveRuleMode, destIp, destPort, forwaredIp, forwardedPort),
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}

	err := removePortCommand.Run()
	fmt.Println(removePortCommand)

	if err != nil {
		log.Print(err)
	}

	return err
}

func iptablesArguments(ruleMode RuleMode, destIp string, destPort int, forwaredIp string, forwardedPort int) []string {

	operationString := "-D"
	if ruleMode == AddRuleMode {
		operationString = "-A"
	}

	return []string{
		ipTablesExecutable,
		"-t",
		"nat",
		operationString,
		"PREROUTING",
		"-d",
		destIp + "/32",
		"-p",
		"tcp",
		"-m",
		"tcp",
		"--dport",
		strconv.Itoa(destPort),
		"-j",
		"DNAT",
		"--to-destination",
		forwaredIp + ":" + strconv.Itoa(forwardedPort),
	}
}
