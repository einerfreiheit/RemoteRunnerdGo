package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type config struct {
	network string
	address string
	timeout int
}

func makeConfig() (c *config) {
	c = new(config)
	flag.StringVar(&c.network, "p", "tcp", "Specify network protocol. Default is tcp ( tcp, tcp4, tcp6, unix or unixpacket are also supported)")
	flag.StringVar(&c.address, "a", ":8081", "Specify address. Default is :8081")
	flag.IntVar(&c.timeout, "t", 1, "Task execution timeout, >= 1 sec. Default 1sec")
	if c.timeout < 1 {
		c.timeout = 1
	}
	return c
}

func main() {
	permissioner := NewPermissioner()
	permissioner.Read([]byte("ping"))

	config := makeConfig()
	server := NewTaskServer(config, permissioner)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	up := make(chan os.Signal, 1)
	signal.Notify(up, syscall.SIGHUP)

	go server.Serve()

	for {
		select {
		case <-up:
			permissioner.Read([]byte("ping"))
		case <-stop:
			log.Println("Shutdown RemoteRunnerd ...")
			server.Shutdown()
			log.Println("RemoteRunnerd is down")
			return
		}
	}
}
