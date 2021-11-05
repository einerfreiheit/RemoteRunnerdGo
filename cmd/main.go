package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const (
	path string = "/etc/remote-runnerd.conf"
)

func read() (c *Config) {
	c = new(Config)
	flag.StringVar(&c.network, "p", "tcp", "Specify network protocol. Default is tcp ( tcp, tcp4, tcp6, unix or unixpacket are also supported)")
	flag.StringVar(&c.address, "a", ":8081", "Specify address. Default is :8081")
	flag.IntVar(&c.timeout, "t", 1, "Specify task execution timeout, sec. Default is 1 sec. Can not be lesser than 1 sec.")
	if c.timeout < 1 {
		c.timeout = 1
	}
	flag.Parse()
	return c
}

func fetch(perm PermissionReaderChecker) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("Failed to read file: ", path)
	}
	perm.Read(content)
}

func main() {
	permissioner := NewPermissioner()
	fetch(permissioner)

	config := read()
	server := NewTaskRunner(config, permissioner)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	up := make(chan os.Signal, 1)
	signal.Notify(up, syscall.SIGHUP)

	go server.Serve()

	for {
		select {
		case <-up:
			fetch(permissioner)
		case <-stop:
			log.Println("Shutdown RemoteRunnerd ...")
			server.Shutdown()
			log.Println("RemoteRunnerd is down")
			return
		}
	}
}
