package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Server interface {
	Handle(conn io.ReadWriter) (err error)
	Serve(address string, network string) (err error)
	Shutdown() (err error)
}

type Runner interface {
	Run(cmd []string, conn io.Writer) (err error)
}

type ServerRunner interface {
	Server
	Runner
}

type Config struct {
	network string
	address string
	timeout int
}

type TaskRunner struct {
	c        *Config
	perm     *permissioner
	quit     chan interface{}
	wg       sync.WaitGroup
	listener net.Listener
}

func (runner *TaskRunner) Serve() {
	defer runner.wg.Done()
	for {
		conn, err := runner.listener.Accept()
		if err != nil {
			select {
			case <-runner.quit:
				log.Println("Stop serving")
				return
			default:
				log.Println("Failed to accept, reason:", err)
				continue
			}
		}
		log.Println("Accepted new client: ", conn.RemoteAddr())
		runner.wg.Add(1)
		go func() {
			defer conn.Close()
			defer runner.wg.Done()
			if runner.Handle(conn) != nil {
				log.Println("Failed to handle incomming client, reason:", err)
			}
		}()
	}
}

func (runner *TaskRunner) Run(command []string, conn io.Writer) (err error) {
	if !runner.perm.Check(command) {
		message := fmt.Sprintf("Command %s is not allowed", strings.Join(command, " "))
		log.Println(message)
		conn.Write([]byte(message))
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(runner.c.timeout)*time.Second)
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
	transfer(stdout, conn)
	transfer(stderr, conn)
	return nil
}

func (runner *TaskRunner) Handle(conn io.ReadWriter) (err error) {
	commands, err := parse(conn)
	if err != nil {
		log.Println("Failed to parse request, reason: ", err)
		return err
	}
	return runner.Run(commands, conn)
}

func (runner *TaskRunner) Shutdown() {
	close(runner.quit)
	runner.listener.Close()
	runner.wg.Wait()
}

func NewTaskRunner(c *Config, perm *permissioner) (runner *TaskRunner) {
	s := &TaskRunner{c: c, perm: perm, quit: make(chan interface{})}
	fmt.Printf("Launching server: %s - %s \n", s.c.network, s.c.address)
	ln, err := net.Listen(s.c.network, s.c.address)
	check(err)
	s.wg.Add(1)
	s.listener = ln
	return s
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func transfer(r io.Reader, w io.Writer) {
	writer := bufio.NewWriter(w)
	n, err := writer.ReadFrom(r)
	log.Println("Sent: ", n)
	if err != nil {
		log.Println("Error while transfer: ", err)
	}
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
