package permission

import (
	"strings"
	"sync"
)

// Reader is an interface for permission reader.
type Reader interface {
	// Update permission with a space-separated array of allowed commands.
	Read([]byte)
}

// Checker is an interface for permission checker.
type Checker interface {
	// Check if command with args (represented by []string) permitted to execute.
	Check([]string) bool
}

// ReaderChecker is an interface for both Reader and Checker.
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
		permitted[strings.Trim(cmd, "\n")] = true
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

// NewPermissioner creates the instance of thread-safe permissioner.
func NewPermissioner() ReaderChecker {
	return &validator{permitted: make(map[string]bool), lock: new(sync.RWMutex)}
}
