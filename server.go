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

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type Server interface {
	Handle(conn io.ReadWriter) (err error)
	Serve(address string, network string) (err error)
	Shutdown() (err error)
}

func parseCommand(conn io.Reader) (command []string, err error) {
	reader := bufio.NewReader(conn)
	data, err := reader.ReadBytes('\n')
	if err != nil {
		return make([]string, 0), err
	}
	request := string(data[0 : len(data)-1])
	return strings.Split(request, " "), nil
}

type Runner interface {
	Run(cmd []string, conn io.Writer) (err error)
}

type TaskServer struct {
	c        *config
	perm     *Permissioner
	quit     chan interface{}
	wg       sync.WaitGroup
	listener net.Listener
}

func (runner *TaskServer) Serve() {
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
		runner.wg.Add(1)
		go func() {
			defer conn.Close()
			if runner.Handle(conn) != nil {
				log.Println("Failed to handle incomming client, reason:", err)
			}
			runner.wg.Done()
		}()
	}
}

func transfer(src io.Reader, dst io.Writer) {
	writer := bufio.NewWriter(dst)
	n, err := writer.ReadFrom(src)
	log.Println("Sent: ", n)
	if err != nil {
		log.Println("Error while transfer: ", err)
	}
}

func (runner *TaskServer) Run(command []string, conn io.Writer) (err error) {
	if !runner.perm.Check(command) {
		message := fmt.Sprintf("Command %s is not allowed", strings.Join(command, " "))
		log.Println(message)
		conn.Write([]byte(message))
		return
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
	return
}

func (runner *TaskServer) Handle(conn io.ReadWriter) (err error) {
	commands, err := parseCommand(conn)
	if err != nil {
		log.Println("Failed to parse request, reason: ", err)
	}
	return runner.Run(commands, conn)
}

func (runner *TaskServer) Shutdown() {
	close(runner.quit)
	runner.listener.Close()
	runner.wg.Wait()
}

func NewTaskServer(c *config, perm *Permissioner) (runner *TaskServer) {
	s := &TaskServer{c: c, perm: perm, quit: make(chan interface{})}
	fmt.Printf("Launching server: %s - %s \n", s.c.network, s.c.address)
	ln, err := net.Listen(s.c.network, s.c.address)
	checkError(err)
	s.wg.Add(1)
	s.listener = ln
	return s
}
