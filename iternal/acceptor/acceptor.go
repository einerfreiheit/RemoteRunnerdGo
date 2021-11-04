package acceptor

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type Acceptor interface {
	AcceptFunc(onAccept func(conn io.ReadWriter) error)
	Serve()
	Stop() error
}

type taskAcceptor struct {
	quit     chan interface{}
	wg       sync.WaitGroup
	onAccept func(conn io.ReadWriter) error
	listener net.Listener
}

func (ts *taskAcceptor) Serve() {
	defer ts.wg.Done()
	for {
		conn, err := ts.listener.Accept()
		if err != nil {
			select {
			case <-ts.quit:
				log.Println("Stop serving")
				return
			default:
				log.Println("Failed to accept, reason:", err)
				continue
			}
		}
		log.Println("Accepted new client: ", conn.RemoteAddr())
		ts.wg.Add(1)
		go func() {
			defer conn.Close()
			defer ts.wg.Done()
			if ts.onAccept(conn) != nil {
				log.Println("Failed to handle incomming client, reason:", err)
			}
		}()
	}
}

func (ts *taskAcceptor) AcceptFunc(onAccept func(conn io.ReadWriter) error) {
	ts.onAccept = onAccept
}

func (ts *taskAcceptor) Stop() error {
	close(ts.quit)
	if err := ts.listener.Close(); err != nil {
		return err
	}
	ts.wg.Wait()
	return nil
}

func NewAcceptor(network string, address string) (Acceptor, error) {
	fmt.Printf("Launching server: network - %s, address - %s \n", network, address)
	ts := &taskAcceptor{quit: make(chan interface{})}
	ln, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	ts.wg.Add(1)
	ts.listener = ln
	return ts, nil
}
