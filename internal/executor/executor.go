package executor

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/einerfreiheit/RemoteRunnerdGo/internal/acceptor"
	"github.com/einerfreiheit/RemoteRunnerdGo/internal/permission"
	"github.com/einerfreiheit/RemoteRunnerdGo/internal/runner"
)

// Executor is an interface for task execution service.
type Executor interface {
	// Start to accept and execute incoming requests.
	Start()
	// Stop accepting requests.
	Stop() error
	// Set task timeout.
	Timeout(time.Duration)
}

type taskExecutor struct {
	ac      acceptor.Acceptor
	timeout time.Duration
	checker permission.Checker
}

func (te *taskExecutor) Start() {
	te.ac.AcceptFunc(te.onAccept)
	te.ac.Serve()
}

func (te *taskExecutor) Stop() error {
	return te.ac.Stop()
}

func (te *taskExecutor) Timeout(timeout time.Duration) {
	te.timeout = timeout
}

// MakeExecutor creates new instance of the Executor.
// Provide protocol, address and permission.Checker to create an exector of remote requests.
// The network must be "tcp", "tcp4", "tcp6", "unix" or "unixpacket".
func MakeExecutor(network string, address string, checker permission.Checker) (e Executor, err error) {
	ac, err := acceptor.NewAcceptor(network, address)
	if err != nil {
		return nil, err
	}
	return &taskExecutor{ac: ac, checker: checker}, nil
}

func (te *taskExecutor) onAccept(conn io.ReadWriter) error {
	cmd, err := parse(conn)
	if err != nil {
		log.Println("Failed to parse request, reason: ", err)
		return err
	}
	if !te.checker.Check(cmd) {
		message := fmt.Sprintf("Command %s is not allowed", strings.Join(cmd, " "))
		log.Println(message)
		conn.Write([]byte(message))
		return nil
	}
	runner := runner.NewTaskRunner(conn)
	return runner.Run(cmd, te.timeout)
}

func parse(conn io.Reader) (command []string, err error) {
	reader := bufio.NewReader(conn)
	data, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	request := string(data)
	return strings.Split(request, " "), nil
}
