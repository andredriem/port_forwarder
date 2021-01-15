package redirect

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"port_forwarder/port_allocation_pool"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator"
)

var validate *validator.Validate
var pkillExecutable (string)
var nodeExecutable (string)

func init() {
	validate = validator.New()
	validate.RegisterValidation("ttlCustomValidator", ttlCustomValidator)
	var err error
	pkillExecutable, err = exec.LookPath("pkill")
	if err != nil {
		panic(err)
	}
	nodeExecutable, err = exec.LookPath("nodejs")
	if err != nil {
		panic(err)
	}
}

type Redirect struct {
	DestIp        string `json:"destIp" validate:"ipv4"`
	DestPort      int    `json:"destPort" validate:"gte=1,lte=65535"`
	ForwaredIp    string `json:"forwaredIp" validate:"ipv4"`
	ForwardedPort int    `json:"forwardedPort" validate:"gte=1,lte=65535"`
	TtlInSeconds  int    `json:"ttlInSeconds" validate:"ttlCustomValidator"`
}

type RuleMode int

const (
	AddRuleMode RuleMode = iota
	RemoveRuleMode
)

func NewRedirect(destIp string, destPort int, forwaredIp string, forwardedPort int, ttlInSeconds int) (*Redirect, error) {

	aRedirect := &Redirect{
		DestIp:        destIp,
		DestPort:      destPort,
		ForwaredIp:    forwaredIp,
		ForwardedPort: forwardedPort,
		TtlInSeconds:  ttlInSeconds,
	}
	// validates
	err := validate.Struct(aRedirect)

	return aRedirect, err
}

func NewRedirectFromJson(httpBody io.Reader) (*Redirect, error) {

	var aRedirect Redirect
	var err error

	_ = json.NewDecoder(httpBody).Decode(&aRedirect)
	err = validate.Struct(&aRedirect)

	if err != nil {
		return nil, err
	}

	return &aRedirect, nil
}

func (redirect *Redirect) AddRedirectToFirewall(portAllocationPool *port_allocation_pool.PortAllocationPool) error {
	addPortCommand := exec.Cmd{
		Path:   nodeExecutable,
		Args:   redirect.proxyServerArguments(),
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}

	err := addPortCommand.Start()
	fmt.Println(addPortCommand)

	if err != nil {
		log.Print(err)
	}

	duration, _ := time.ParseDuration(strconv.Itoa(redirect.TtlInSeconds) + "s")

	time.AfterFunc(duration, func() {
		redirect.RemoveRedirectFromFirewall(portAllocationPool)
	})

	return err
}

func (redirect *Redirect) RemoveRedirectFromFirewall(portAllocationPool *port_allocation_pool.PortAllocationPool) error {
	removePortCommand := exec.Cmd{
		Path:   pkillExecutable,
		Args:   append([]string{pkillExecutable, "-f"}, strings.Join(redirect.proxyServerArguments(), " ")),
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}

	err := removePortCommand.Run()
	fmt.Println(removePortCommand)

	if err != nil {
		log.Print(err)
	}

	portAllocationPool.DeallocatePort(redirect.ForwardedPort)

	return err
}

func (redirect *Redirect) proxyServerArguments() []string {
	return []string{
		nodeExecutable,
		"jsproxy.js",
		strconv.Itoa(redirect.ForwardedPort),
		redirect.DestIp,
		strconv.Itoa(redirect.DestPort),
	}

}

func ttlCustomValidator(fl validator.FieldLevel) bool {

	ttlValue := fl.Field().Int()

	if ttlValue == -1 {
		return true
	}

	if ttlValue > 0 {
		return true
	}

	return false
}
