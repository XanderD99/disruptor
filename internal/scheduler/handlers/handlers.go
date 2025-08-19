package handlers

import (
	"sync"
)

// Generic singleton registry for handlers
var (
	handlerRegistry = make(map[string]any)
	handlerOnce     = make(map[string]*sync.Once)
	registryMu      sync.Mutex
)

// Register a builder for a handler type
func registerHandlerSingleton(key string, builder func() any) {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, exists := handlerOnce[key]; !exists {
		handlerOnce[key] = &sync.Once{}
	}
	handlerOnce[key].Do(func() {
		handlerRegistry[key] = builder()
	})
}

// Get the singleton handler for a key
func getHandlerSingleton(key string) any {
	registryMu.Lock()
	defer registryMu.Unlock()
	return handlerRegistry[key]
}
