package runner

import (
	"bufio"
	"context"
	"io"
	"log"
	"os/exec"
	"time"
)

// Runner is an task runner interface
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
	transfer(stdout, tr.w)
	transfer(stderr, tr.w)
	return nil
}

func transfer(r io.Reader, w io.Writer) {
	writer := bufio.NewWriter(w)
	n, err := writer.ReadFrom(r)
	log.Println("Sent: ", n)
	if err != nil {
		log.Println("Error while transfer: ", err)
	}
}

// NewTaskRunner creates Runner instance. Provide io.Writer for task output.
func NewTaskRunner(w io.Writer) Runner {
	return &taskRunner{w: w}
}
