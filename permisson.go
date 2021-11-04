package main

import (
	"strings"
	"sync"
)

type PermissionChecker interface {
	Check(cmds []string) (allowed bool)
}

type PermissionReader interface {
	Read(data []byte)
}

type PermissionReaderChecker interface {
	PermissionReader
	PermissionChecker
}

type Permissioner struct {
	permitted map[string]bool
	lock      *sync.RWMutex
}

func (perm *Permissioner) Read(data []byte) {
	commands := strings.Split(string(data), " ")
	permitted := make(map[string]bool)
	for _, cmd := range commands {
		permitted[cmd] = true
	}
	perm.lock.Lock()
	defer perm.lock.Unlock()
	perm.permitted = permitted
}

func (perm *Permissioner) Check(commands []string) (allowed bool) {
	if len(commands) == 0 {
		return false
	}
	for _, cmd := range commands {
		if strings.Contains(cmd, "&") {
			return false
		}
	}
	perm.lock.RLock()
	defer perm.lock.RUnlock()
	if _, ok := perm.permitted[commands[0]]; ok {
		return true
	}
	return false
}

func NewPermissioner() (perm *Permissioner) {
	return &Permissioner{permitted: make(map[string]bool), lock: new(sync.RWMutex)}
}
