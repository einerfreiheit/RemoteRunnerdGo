# RemoteRunnerdGo
[![Go Report Card](https://goreportcard.com/badge/github.com/einerfreiheit/RemoteRunnerdGo)](https://goreportcard.com/report/github.com/einerfreiheit/RemoteRunnerdGo)
[![Go Reference](https://pkg.go.dev/badge/github.com/einerfreiheit/RemoteRunnerdGo.svg)](https://pkg.go.dev/github.com/einerfreiheit/RemoteRunnerdGo)


Simple remote task runner. Runner executes requests and sends result back. Data can be transmitted via TCP or UDS (Unix Domain Socket). 

 - Configuration:

       /etc/remote-runnerd.conf - space separated list of permitted commands.
      
 - Options:

       -t: task execution timeout, sec; default: 0.
    
       -a: address (port for TCP, path for UDS)
       
       -p: protocol (tcp, tcp4, tcp6, unix)
        
  - Usage example:
  
         ./RemoteRunnerdGo -t 10 -p tcp -a :12345
 
 - Build:
 
       git clone https://github.com/einerfreiheit/RemoteRunnerdGo.git
       go build
