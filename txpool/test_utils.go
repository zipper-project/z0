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

package txpool

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/zipper-project/z0/common"
	"github.com/zipper-project/z0/core/asset"
	"github.com/zipper-project/z0/crypto"
	"github.com/zipper-project/z0/feed"
	"github.com/zipper-project/z0/params"
	"github.com/zipper-project/z0/state"
	"github.com/zipper-project/z0/types"
	"github.com/zipper-project/z0/utils/zdb"
)

const amount = 1e18

// testTxPoolConfig is a transaction pool configuration without stateful disk
// sideeffects used during testing.
var testTxPoolConfig Config

func init() {
	testTxPoolConfig = Config{
		Journal:   "",
		Rejournal: time.Hour,

		PriceLimit: 1,
		PriceBump:  10,

		AccountSlots: 16,
		GlobalSlots:  4096,
		AccountQueue: 64,
		GlobalQueue:  1024,

		Lifetime: 3 * time.Hour,
	}
}

type testBlockChain struct {
	statedb       *state.StateDB
	gasLimit      uint64
	chainHeadFeed *feed.Feed
}

func (bc *testBlockChain) CurrentBlock() *types.Block {
	return types.NewBlock(&types.Header{
		GasLimit: bc.gasLimit,
	}, nil, nil, nil)
}

func (bc *testBlockChain) GetBlock(hash common.Hash, number uint64) *types.Block {
	return bc.CurrentBlock()
}

func (bc *testBlockChain) StateAt(common.Hash) (*state.StateDB, error) {
	return bc.statedb, nil
}

func (bc *testBlockChain) SubscribeChainHeadEvent(ch chan<- ChainHeadEvent) feed.Subscription {
	return bc.chainHeadFeed.Subscribe(ch)
}

func transaction(nonce uint64, gaslimit uint64, key *ecdsa.PrivateKey) *types.Transaction {
	return pricedTransaction(nonce, gaslimit, big.NewInt(1), key)
}

func pricedTransaction(nonce uint64, gaslimit uint64, gasprice *big.Int, key *ecdsa.PrivateKey) *types.Transaction {
	tx, _ := types.SignTx(newTx(nonce, big.NewInt(100), gaslimit, gasprice, nil), types.NewSigner(params.DefaultChainconfig.ChainID), key)
	return tx
}

func setupTxPool() (*TxPool, *ecdsa.PrivateKey) {
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(zdb.NewMemDatabase()))
	blockchain := &testBlockChain{statedb, 1000000, new(feed.Feed)}

	key, _ := crypto.GenerateKey()
	pool := New(testTxPoolConfig, params.DefaultChainconfig, blockchain)

	return pool, key
}

// validateTxPoolInternals checks various consistency invariants within the pool.
func validateTxPoolInternals(pool *TxPool) error {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	// Ensure the total transaction set is consistent with pending + queued
	pending, queued := pool.stats()
	if total := pool.all.Count(); total != pending+queued {
		return fmt.Errorf("total transaction count %d != %d pending + %d queued", total, pending, queued)
	}
	if priced := pool.priced.items.Len() - pool.priced.stales; priced != pending+queued {
		return fmt.Errorf("total priced transaction count %d != %d pending + %d queued", priced, pending, queued)
	}
	// Ensure the next nonce to assign is the correct one

	for addr, list := range pool.pending {
		// Find the last transaction
		var last uint64
		for nonce := range list.txs.items {
			if last < nonce {
				last = nonce
			}
		}

		if nonce := pool.pendingAsset.GetNonce(addr); nonce != last+1 {
			return fmt.Errorf("pending nonce mismatch: have %v, want %v", nonce, last+1)
		}
	}
	return nil
}

// validateEvents checks that the correct number of transaction addition events
// were fired on the pool's event feed.
func validateEvents(events chan NewTxsEvent, count int) error {
	var received []*types.Transaction

	for len(received) < count {
		select {
		case ev := <-events:
			received = append(received, ev.Txs...)
		case <-time.After(time.Second):
			return fmt.Errorf("event #%d not fired", received)
		}
	}
	if len(received) > count {
		return fmt.Errorf("more than %d events fired: %v", count, received[count:])
	}
	select {
	case ev := <-events:
		return fmt.Errorf("more than %d events fired: %v", count, ev.Txs)

	case <-time.After(50 * time.Millisecond):
		// This branch should be "default", but it's a data race between goroutines,
		// reading the event channel and pushing into it, so better wait a bit ensuring
		// really nothing gets injected.
	}
	return nil
}

func deriveSender(tx *types.Transaction) (common.Address, error) {
	return types.Sender(types.NewSigner(params.DefaultChainconfig.ChainID), tx)
}

type testChain struct {
	*testBlockChain

	address common.Address
	trigger *bool
}

// testChain.State() is used multiple times to reset the pending state.
// when simulate is true it will create a state that indicates
// that tx0 and tx1 are included in the chain.
func (c *testChain) State() (*state.StateDB, error) {
	// delay "state change" by one. The tx pool fetches the
	// state multiple times and by delaying it a bit we simulate
	// a state change between those fetches.
	stdb := c.statedb

	if *c.trigger {
		c.statedb, _ = state.New(common.Hash{}, state.NewDatabase(zdb.NewMemDatabase()))
		assert := asset.NewAsset(c.statedb)

		// simulate that the new head block included tx0 and tx1
		assert.SetNonce(c.address, 2)
		assert.AddBalance(c.address, common.Address{}, new(big.Int).SetUint64(amount))
		*c.trigger = false
	}
	return stdb, nil
}

func newTx(nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *types.Transaction {
	new := types.NewTransaction(nonce, gasLimit, gasPrice, data)
	new.WithInput(types.AMInput{AssertID: &types.ZipAssetID, Payload: nil})
	new.WithOutput(types.AMOutput{AssertID: &types.ZipAssetID, Address: &common.Address{}, Value: amount})
	return new
}
