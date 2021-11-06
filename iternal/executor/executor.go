package executor

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/einerfreiheit/RemoteRunnerdGo/iternal/acceptor"
	"github.com/einerfreiheit/RemoteRunnerdGo/iternal/permission"
	"github.com/einerfreiheit/RemoteRunnerdGo/iternal/runner"
)

// Interface of a task executor
type Executor interface {
	// Start to accept and execute incoming requests
	Start()
	// Stop accepting requests
	Stop() error
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
		return make([]string, 0), err
	}
	request := string(data[0 : len(data)-1])
	return strings.Split(request, " "), nil
}
