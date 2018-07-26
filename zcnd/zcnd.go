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
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/log"
	"github.com/zipper-project/z0/core"
	"github.com/zipper-project/z0/node"
	"github.com/zipper-project/z0/params"
	"github.com/zipper-project/z0/rawdb"
	"github.com/zipper-project/z0/rpc"
	"github.com/zipper-project/z0/txpool"
	"github.com/zipper-project/z0/zdb"
)

// Zcnd implements the z0 service.
type Zcnd struct {
	config       *Config
	chainConfig  *params.ChainConfig
	shutdownChan chan bool // Channel for shutting down the service
	blockchain   *core.BlockChain
	txPool       *txpool.TxPool
	chainDb      zdb.Database // Block chain database

	lock sync.RWMutex // Protects the variadic fields (e.g. gas price)
}

// New creates a new Zcnd object (including the
// initialisation of the common Zcnd object)
func New(ctx *node.ServiceContext, config *Config) (*Zcnd, error) {
	cfg, err := json.Marshal(config)
	log.Info("znd config :", "config", string(cfg))

	chainDb, err := CreateDB(ctx, config, "chaindata")
	if err != nil {
		return nil, err
	}

	chainCfg, _, err := core.SetupGenesisBlock(chainDb, config.Genesis)
	if err != nil {
		return nil, err
	}
	log.Info("Initialised chain configuration", "config", chainCfg)

	zcnd := &Zcnd{
		config:       config,
		chainDb:      chainDb,
		chainConfig:  chainCfg,
		shutdownChan: make(chan bool),
	}

	if !config.SkipBcVersionCheck {
		bcVersion := rawdb.ReadDatabaseVersion(chainDb)
		if bcVersion != core.BlockChainVersion && bcVersion != 0 {
			return nil, fmt.Errorf("Blockchain DB version mismatch (%d / %d)", bcVersion, core.BlockChainVersion)
		}
		rawdb.WriteDatabaseVersion(chainDb, core.BlockChainVersion)
	}
	cacheConfig := &core.CacheConfig{Disabled: config.NoPruning, TrieNodeLimit: config.TrieCache, TrieTimeLimit: config.TrieTimeout}
	//blockchain
	zcnd.blockchain, err = core.New(chainDb, cacheConfig, zcnd.chainConfig)
	if err != nil {
		return nil, err
	}

	// txpool
	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}

	// todo add blockchian
	zcnd.txPool = txpool.New(config.TxPool, zcnd.chainConfig, zcnd.blockchain)

	return zcnd, nil
}

// APIs return the collection of RPC services the zcnd package offers.
func (z *Zcnd) APIs() []rpc.API { return nil }

// Start implements node.Service, starting all internal goroutines.
func (z *Zcnd) Start() error {
	log.Info("start zcnd...")
	return nil
}

// Stop implements node.Service, terminating all internal goroutine
func (z *Zcnd) Stop() error {
	z.txPool.Stop()
	z.chainDb.Close()
	close(z.shutdownChan)
	return nil
}

// CreateDB creates the chain database.
func CreateDB(ctx *node.ServiceContext, config *Config, name string) (zdb.Database, error) {
	db, err := ctx.OpenDatabase(name, config.DatabaseCache, config.DatabaseHandles)
	if err != nil {
		return nil, err
	}
	return db, nil
}
