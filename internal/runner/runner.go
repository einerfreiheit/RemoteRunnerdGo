package runner

import (
	"bufio"
	"context"
	"io"
	"log"
	"os/exec"
	"time"
)

// Runner is an task runner interface.
type Runner interface {
	// Execute a cmd with args (represented by []string) with specified timeout (sec).
	Run(cmd []string, timeout time.Duration) error
}

type taskRunner struct {
	w io.Writer
}

func (tr *taskRunner) Run(command []string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}
	err = transfer(stdout, tr.w, "stdout")
	if err != nil {
		return err
	}
	return transfer(stderr, tr.w, "stderr")
}

func transfer(r io.Reader, w io.Writer, source string) error {
	writer := bufio.NewWriter(w)
	n, err := writer.ReadFrom(r)
	log.Printf("Sent %d bytes from %s\n", n, source)
	return err
}

// NewTaskRunner creates the Runner instance. Provide io.Writer for task output.
func NewTaskRunner(w io.Writer) Runner {
	return &taskRunner{w: w}
}
