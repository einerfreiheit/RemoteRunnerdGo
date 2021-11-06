package permission

import (
	"strings"
	"sync"
)

// Interface for permission reader
type Reader interface {
	// Update permission
	Read([]byte)
}

// Interface for permission checker
type Checker interface {
	// Check commands permission
	Check([]string) bool
}

type ReaderChecker interface {
	Reader
	Checker
}

type validator struct {
	permitted map[string]bool
	lock      *sync.RWMutex
}

func (v *validator) Read(data []byte) {
	commands := strings.Split(string(data), " ")
	permitted := make(map[string]bool)
	for _, cmd := range commands {
		permitted[cmd] = true
	}
	v.lock.Lock()
	defer v.lock.Unlock()
	v.permitted = permitted
}

func (v *validator) Check(commands []string) bool {
	if len(commands) == 0 {
		return false
	}
	for _, cmd := range commands {
		if strings.Contains(cmd, "&") {
			return false
		}
	}
	v.lock.RLock()
	defer v.lock.RUnlock()
	if _, ok := v.permitted[commands[0]]; ok {
		return true
	}
	return false
}

// Create the permissioner instance.
// Permissioner is thread-safe.
func NewPermissioner() ReaderChecker {
	return &validator{permitted: make(map[string]bool), lock: new(sync.RWMutex)}
}
