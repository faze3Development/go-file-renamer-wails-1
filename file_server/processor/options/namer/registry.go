package namer

import (
	"fmt"
	"sync"
)

var (
	namerRegistry = make(map[string]Namer)
	mu            sync.RWMutex
)

// RegisterNamer adds a namer to the global registry.
// This is intended to be called from the init() function of each namer implementation.
func RegisterNamer(namer Namer) {
	mu.Lock()
	defer mu.Unlock()
	id := namer.Info().ID
	if _, exists := namerRegistry[id]; exists {
		panic(fmt.Sprintf("Namer with ID '%s' is already registered", id))
	}
	namerRegistry[id] = namer
}

// GetNamerInfo returns information for all registered namers.
func GetNamerInfo() []Info {
	mu.RLock()
	defer mu.RUnlock()
	var info []Info
	for _, namer := range namerRegistry {
		info = append(info, namer.Info())
	}
	return info
}
