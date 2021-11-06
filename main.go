package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/einerfreiheit/RemoteRunnerdGo/iternal/executor"
	"github.com/einerfreiheit/RemoteRunnerdGo/iternal/permission"
)

const (
	path string = "/etc/remote-runnerd.conf"
)

type config struct {
	network string
	address string
	timeout time.Duration
}

func read() (c *config) {
	c = new(config)
	flag.StringVar(&c.network, "p", "tcp", "Specify network protocol. Default is tcp ( tcp, tcp4, tcp6, unix or unixpacket are also supported)")
	flag.StringVar(&c.address, "a", ":8081", "Specify address. Default is :8081")
	var timeout int
	flag.IntVar(&timeout, "t", 1, "Specify task execution timeout, sec. Default is 1 sec. Can not be lesser than 1 sec.")
	flag.Parse()
	if timeout < 1 {
		timeout = 1
	}
	c.timeout = time.Duration(timeout * int(time.Second))
	return c
}

func fetch(perm permission.Reader) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("Failed to read file: ", path)
	}
	perm.Read(content)
}

func main() {
	permissioner := permission.NewPermissioner()
	fetch(permissioner)

	config := read()
	executor, err := executor.MakeExecutor(config.network, config.address, permissioner)
	if err != nil {
		log.Fatal(err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	up := make(chan os.Signal, 1)
	signal.Notify(up, syscall.SIGHUP)

	go executor.Start()
	for {
		select {
		case <-up:
			fetch(permissioner)
		case <-stop:
			log.Println("Shutdown runnerd ...")
			if err = executor.Stop(); err != nil {
				log.Fatal(err)
			}
			log.Println("runnerd is down")
			return
		}
	}
}
