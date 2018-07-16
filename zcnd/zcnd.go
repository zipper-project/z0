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

package zcnd

import (
	"sync"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/log"
	"github.com/zipper-project/z0/node"
	"github.com/zipper-project/z0/rpc"
	"github.com/zipper-project/z0/zdb"
)

// Zcnd implements the z0 service.
type Zcnd struct {
	config       *Config
	shutdownChan chan bool // Channel for shutting down the service
	txPool       *core.TxPool
	chainDb      zdb.Database // Block chain database

	lock sync.RWMutex // Protects the variadic fields (e.g. gas price)
}

// New creates a new Zcnd object (including the
// initialisation of the common Zcnd object)
func New(ctx *node.ServiceContext, config *Config) (*Zcnd, error) {
	return nil, nil
}

// APIs return the collection of RPC services the zcnd package offers.
func (z *Zcnd) APIs() []rpc.API { return nil }

// Start implements node.Service, starting all internal goroutines.
func (z *Zcnd) Start() error {
	log.Info("start zcnd...")
	return nil
}

// Stop implements node.Service, terminating all internal goroutine
func (z *Zcnd) Stop() error { return nil }
