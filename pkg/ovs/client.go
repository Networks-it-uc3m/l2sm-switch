package ovs

import (
	"bytes"
	"os/exec"
)

// CommandExecutor defines an interface for running external commands.
// This allows for mocking the execution of commands like 'ovs-vsctl' or 'ip' during testing.
type Client interface {
	CombinedOutput(args ...string) ([]byte, error)
	Run(args ...string) error
	Output(args ...string) ([]byte, error)
	OutputToBuffer(stdout *bytes.Buffer, args ...string) error
}

type ClientCommand string

const (
	OvsVsctlClient ClientCommand = "ovs-vsctl"
	IpClient       ClientCommand = "ip"
)

// DefaultClient is the standard implementation that uses os/exec.
type DefaultClient struct {
	command string
	sudo    bool
}

func (e *DefaultClient) buildCommand(args ...string) *exec.Cmd {
	if e.sudo {
		fullArgs := append([]string{e.command}, args...)
		return exec.Command("sudo", fullArgs...)
	}
	return exec.Command(e.command, args...)
}

func (e *DefaultClient) CombinedOutput(args ...string) ([]byte, error) {
	cmd := e.buildCommand(args...)
	return cmd.CombinedOutput()
}

func (e *DefaultClient) Run(args ...string) error {
	cmd := e.buildCommand(args...)
	return cmd.Run()
}

func (e *DefaultClient) Output(args ...string) ([]byte, error) {
	cmd := e.buildCommand(args...)
	return cmd.Output()
}

func (e *DefaultClient) OutputToBuffer(stdout *bytes.Buffer, args ...string) error {
	cmd := e.buildCommand(args...)
	cmd.Stdout = stdout
	return cmd.Run()
}

// NewClient creates a new instance of the default command executor with optional sudo.
func NewClient(command ClientCommand) Client {
	return &DefaultClient{command: string(command), sudo: false}
}

// NewClient creates a new instance of the default command executor with optional sudo.
func NewSudoClient(command ClientCommand) Client {
	return &DefaultClient{command: string(command), sudo: true}
}
