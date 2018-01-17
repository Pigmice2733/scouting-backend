package lastmodified

import "sync"

// Manager handles getting and setting lastModified dates on a certain path.
type Manager struct {
	pathLastModified map[string]string
	mu               *sync.Mutex
}

// New returns a new last modified manager.
func New() Manager {
	return Manager{make(map[string]string), new(sync.Mutex)}
}

// Get retrieves the last modified string for a given URL.
func (m Manager) Get(path string) (lastModified string) {
	m.mu.Lock()
	lastModified = m.pathLastModified[path]
	m.mu.Unlock()
	return
}

// Set sets the last modified string for a URL.
func (m Manager) Set(path, lastModified string) {
	m.mu.Lock()
	m.pathLastModified[path] = lastModified
	m.mu.Unlock()
}
