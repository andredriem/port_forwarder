package redirect

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/go-playground/validator"
)

var validate *validator.Validate

type Redirect struct {
	DestIp        string `json:"destIp" validate:"ipv4"`
	DestPort      int    `json:"destPort" validate:"gte=1,lte=4096"`
	ForwaredIp    string `json:"forwaredIp" validate:"ipv4"`
	ForwardedPort int    `json:"forwardedPort" validate:"gte=1,lte=4096"`
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

	return &aRedirect, err
}

var ipTablesExecutable (string)

func AddRedirectToFirewall(redirect *Redirect) error {
	addPortCommand := exec.Cmd{
		Path:   ipTablesExecutable,
		Args:   iptablesArguments(AddRuleMode, redirect),
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}

	err := addPortCommand.Run()
	fmt.Println(addPortCommand)

	if err != nil {
		log.Print(err)
	}

	duration, _ := time.ParseDuration(strconv.Itoa(redirect.TtlInSeconds) + "s")

	time.AfterFunc(duration, func() {
		RemoveRedirectFromFirewall(redirect)
	})

	return err
}

func RemoveRedirectFromFirewall(redirect *Redirect) error {
	removePortCommand := exec.Cmd{
		Path:   ipTablesExecutable,
		Args:   iptablesArguments(RemoveRuleMode, redirect),
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

func iptablesArguments(ruleMode RuleMode, redirect *Redirect) []string {

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
		redirect.DestIp + "/32",
		"-p",
		"tcp",
		"-m",
		"tcp",
		"--dport",
		strconv.Itoa(redirect.DestPort),
		"-j",
		"DNAT",
		"--to-destination",
		redirect.ForwaredIp + ":" + strconv.Itoa(redirect.ForwardedPort),
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

func init() {
	validate = validator.New()
	validate.RegisterValidation("ttlCustomValidator", ttlCustomValidator)
}
