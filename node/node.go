// Copyright 2018 The zipper team Authors
// This file is part of the z0 library.
//
// The z0 library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The z0 library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the z0 library. If not, see <http://www.gnu.org/licenses/>.

package node

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"

	"github.com/ethereum/go-ethereum/log"
	"github.com/zipper-project/z0/utils/filelock"
)

// Node is a container on which services can be registered.
type Node struct {
	config          *Config
	running         bool
	instanceDirLock filelock.Releaser        // prevents concurrent use of instance directory
	serviceFuncs    []ServiceConstructor     // Service constructors (in dependency order)
	services        map[reflect.Type]Service // Currently running services
	stop            chan struct{}            // Channel to wait for termination notifications
	lock            sync.RWMutex

	log log.Logger
}

// New creates a new P2P node, ready for protocol registration.
func New(conf *Config) *Node {
	return &Node{
		config:       conf,
		running:      false,
		serviceFuncs: []ServiceConstructor{},
		services:     make(map[reflect.Type]Service),
		log:          conf.Logger,
	}
}

// Register injects a new service into the node's stack. The service created by
// the passed constructor must be unique in its type with regard to sibling ones.
func (n *Node) Register(constructor ServiceConstructor) error {
	n.lock.Lock()
	defer n.lock.Unlock()
	if n.running {
		return ErrNodeRunning
	}
	n.serviceFuncs = append(n.serviceFuncs, constructor)
	return nil
}

// Start create a live node and starts running it.
func (n *Node) Start() error {
	n.lock.Lock()
	defer n.lock.Unlock()

	if n.running {
		return ErrNodeRunning
	}

	if err := n.openDataDir(); err != nil {
		return err
	}

	services := make(map[reflect.Type]Service)
	for _, constructor := range n.serviceFuncs {
		// Create a new context for the particular service
		ctx := &ServiceContext{
			config:   n.config,
			services: make(map[reflect.Type]Service),
		}
		for kind, s := range services { // copy needed for threaded access
			ctx.services[kind] = s
		}
		// Construct and save the service
		service, err := constructor(ctx)
		if err != nil {
			return err
		}
		kind := reflect.TypeOf(service)
		if _, exists := services[kind]; exists {
			return fmt.Errorf("duplicate service: %v", kind)
		}
		services[kind] = service
	}

	// Start each of the services
	started := []reflect.Type{}
	for kind, service := range services {
		// Start the next service, stopping all previous upon failure
		if err := service.Start(); err != nil {
			for _, kind := range started {
				services[kind].Stop()
			}
			return err
		}
		// Mark the service started for potential cleanup
		started = append(started, kind)
	}

	// todo Lastly start the configured RPC interfaces
	// if err := n.startRPC(services); err != nil {
	// 	for _, service := range services {
	// 		service.Stop()
	// 	}
	// 	return err
	// }
	n.services = services
	n.running = true
	n.stop = make(chan struct{})
	return nil
}

// Stop terminates a running node along with all it's services. In the node was
// not started, an error is returned.
func (n *Node) Stop() error {
	n.lock.Lock()
	defer n.lock.Unlock()

	// Short circuit if the node's not running
	if !n.running {
		return ErrNodeStopped
	}

	failure := &StopError{
		Services: make(map[reflect.Type]error),
	}
	for kind, service := range n.services {
		if err := service.Stop(); err != nil {
			failure.Services[kind] = err
		}
	}
	n.services = nil

	n.releaseInstanceDir()

	close(n.stop)

	n.running = false

	if len(failure.Services) > 0 {
		return failure
	}

	return nil
}

// Wait blocks the thread until the node is stopped. If the node is not running
// at the time of invocation, the method immediately returns.
func (n *Node) Wait() {
	n.lock.RLock()
	if !n.running {
		n.lock.RUnlock()
		return
	}
	stop := n.stop
	n.lock.RUnlock()

	<-stop
}

// Restart terminates a running node and boots up a new one in its place. If the
// node isn't running, an error is returned.
func (n *Node) Restart() error {
	if err := n.Stop(); err != nil {
		return err
	}
	return n.Start()
}

// Service retrieves a currently running service registered of a specific type.
func (n *Node) Service(service interface{}) error {
	n.lock.RLock()
	defer n.lock.RUnlock()

	// Short circuit if the node's not running
	if !n.running {
		return ErrNodeStopped
	}
	// Otherwise try to find the service to return
	element := reflect.ValueOf(service).Elem()
	if running, ok := n.services[element.Type()]; ok {
		element.Set(reflect.ValueOf(running))
		return nil
	}
	return ErrServiceUnknown
}

func (n *Node) openDataDir() error {
	if n.config.DataDir == "" {
		return nil
	}

	instdir := filepath.Join(n.config.DataDir, n.config.Name)
	if err := os.MkdirAll(instdir, 0700); err != nil {
		return err
	}
	return n.lockInstanceDir(instdir)
}

// lockInstanceDir Lock the instance directory to prevent concurrent use by another instance as well as
// accidental use of the instance directory as a database.
func (n *Node) lockInstanceDir(path string) error {
	release, _, err := filelock.New(filepath.Join(path, "LOCK"))
	if err != nil {
		return convertFileLockError(err)
	}
	n.instanceDirLock = release
	return nil
}

// releaseInstanceDir Release instance directory lock.
func (n *Node) releaseInstanceDir() {
	if n.instanceDirLock != nil {
		if err := n.instanceDirLock.Release(); err != nil {
			n.log.Error("Can't release datadir lock", "err", err)
		}
		n.instanceDirLock = nil
	}
}
