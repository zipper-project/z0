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
	"fmt"
	"math"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/zipper-project/z0/common"
	"github.com/zipper-project/z0/event"
	"github.com/zipper-project/z0/params"
	"github.com/zipper-project/z0/state"
	"github.com/zipper-project/z0/types"
	"gopkg.in/karalabe/cookiejar.v2/collections/prque"
)

var (
	evictionInterval    = time.Minute     // Time interval to check for evictable transactions
	statsReportInterval = 8 * time.Second // Time interval to report transaction pool stats
)

// blockChain provides the state of blockchain and current gas limit to do
// some pre checks in tx pool and event subscribers.
type blockChain interface {
	CurrentBlock() *types.Block
	GetBlock(hash common.Hash, number uint64) *types.Block
	StateAt(root common.Hash) (*state.StateDB, error)

	SubscribeChainHeadEvent(ch chan<- ChainHeadEvent) event.Subscription
}

// TxPool contains all currently known transactions. Transactions
// enter the pool when they are received from the network or submitted
// locally. They exit the pool when they are included in the blockchain.
type TxPool struct {
	config       *Config
	gasPrice     *big.Int
	chain        blockChain
	signer       types.Signer
	mu           sync.RWMutex
	chainHeadCh  chan ChainHeadEvent
	chainHeadSub event.Subscription
	currentState *state.StateDB // Current state in the blockchain head
	//pendingState  *state.ManagedState // Pending state tracking virtual nonces
	currentMaxGas uint64 // Current gas limit for transaction caps

	locals  *accountSet                  // Set of local transaction to exempt from eviction rules
	journal *txJournal                   // Journal of local transaction to back up to disk
	pending map[common.Address]*txList   // All currently processable transactions
	queue   map[common.Address]*txList   // Queued but non-processable transactions
	beats   map[common.Address]time.Time // Last heartbeat from each known account
	all     *txLookup                    // All transactions to allow lookups
	priced  *txPricedList                // All transactions sorted by price

	wg sync.WaitGroup // for shutdown sync
}

// New creates a new transaction pool to gather, sort and filter inbound
// transactions from the network.
func New(config *Config, chainconfig *params.ChainConfig, bc blockChain) *TxPool {
	signer := types.NewSigner(chainconfig.ChainID)
	all := newTxLookup()
	pool := &TxPool{
		config:      config.check(),
		chain:       bc,
		signer:      signer,
		locals:      newAccountSet(signer),
		chainHeadCh: make(chan ChainHeadEvent, chainHeadChanSize),
		pending:     make(map[common.Address]*txList),
		queue:       make(map[common.Address]*txList),
		beats:       make(map[common.Address]time.Time),
		all:         all,
		priced:      newTxPricedList(all),
		gasPrice:    new(big.Int).SetUint64(config.PriceLimit),
	}
	pool.reset(nil, bc.CurrentBlock().Header())

	// If local transactions and journaling is enabled, load from disk
	if !config.NoLocals && config.Journal != "" {
		pool.journal = newTxJournal(config.Journal)
		if err := pool.journal.load(pool.AddLocals); err != nil {
			log.Warn("Failed to load transaction journal", "err", err)
		}
		if err := pool.journal.rotate(pool.local()); err != nil {
			log.Warn("Failed to rotate transaction journal", "err", err)
		}
	}
	// Subscribe events from blockchain
	pool.chainHeadSub = pool.chain.SubscribeChainHeadEvent(pool.chainHeadCh)

	// Start the event loop and return
	pool.wg.Add(1)
	go pool.loop()

	return pool
}

// loop is the transaction pool's main event loop, waiting for and reacting to
// outside blockchain events as well as for various reporting and transaction
// eviction events.
func (tp *TxPool) loop() {
	defer tp.wg.Done()

	// Start the stats reporting and transaction eviction tickers
	var prevPending, prevQueued, prevStales int

	report := time.NewTicker(statsReportInterval)
	defer report.Stop()

	evict := time.NewTicker(evictionInterval)
	defer evict.Stop()

	journal := time.NewTicker(tp.config.Rejournal)
	defer journal.Stop()

	// Track the previous head headers for transaction reorgs
	head := tp.chain.CurrentBlock()

	// Keep waiting for and reacting to the various events
	for {
		select {
		// Handle ChainHeadEvent
		case ev := <-tp.chainHeadCh:
			if ev.Block != nil {
				tp.mu.Lock()
				tp.reset(head.Header(), ev.Block.Header())
				head = ev.Block

				tp.mu.Unlock()
			}
		// Be unsubscribed due to system stopped
		case <-tp.chainHeadSub.Err():
			return

		// Handle stats reporting ticks
		case <-report.C:
			tp.mu.RLock()
			pending, queued := tp.stats()
			stales := tp.priced.stales
			tp.mu.RUnlock()

			if pending != prevPending || queued != prevQueued || stales != prevStales {
				log.Debug("Transaction pool status report", "executable", pending, "queued", queued, "stales", stales)
				prevPending, prevQueued, prevStales = pending, queued, stales
			}

		// Handle inactive account transaction eviction
		case <-evict.C:
			tp.mu.Lock()
			for addr := range tp.queue {
				// Skip local transactions from the eviction mechanism
				if tp.locals.contains(addr) {
					continue
				}
				// Any non-locals old enough should be removed
				if time.Since(tp.beats[addr]) > tp.config.Lifetime {
					for _, tx := range tp.queue[addr].Flatten() {
						tp.removeTx(tx.Hash(), true)
					}
				}
			}
			tp.mu.Unlock()

		// Handle local transaction journal rotation
		case <-journal.C:
			if tp.journal != nil {
				tp.mu.Lock()
				if err := tp.journal.rotate(tp.local()); err != nil {
					log.Warn("Failed to rotate local tx journal", "err", err)
				}
				tp.mu.Unlock()
			}
		}
	}
}

// stats retrieves the current pool stats, namely the number of pending and the
// number of queued (non-executable) transactions.
func (tp *TxPool) stats() (int, int) {
	pending := 0
	for _, list := range tp.pending {
		pending += list.Len()
	}
	queued := 0
	for _, list := range tp.queue {
		queued += list.Len()
	}
	return pending, queued
}

//Stop terminates the transaction pool.
func (tp *TxPool) Stop() {

	// Unsubscribe subscriptions registered from blockchain
	tp.chainHeadSub.Unsubscribe()
	tp.wg.Wait()

	if tp.journal != nil {
		tp.journal.close()
	}
	log.Info("Transaction pool stopped")

}

// reset retrieves the current state of the blockchain and ensures the content
// of the transaction pool is valid with regard to the chain state.
func (tp *TxPool) reset(oldHeader, newHeader *types.Header) {
	var reinject types.Transactions
	if oldHeader != nil && oldHeader.Hash() != newHeader.ParentHash {
		oldNum := oldHeader.Number.Uint64()
		newNum := newHeader.Number.Uint64()
		if depth := uint64(math.Abs(float64(oldNum) - float64(newNum))); depth > 64 {
			log.Debug("Skipping deep transaction reorg", "depth", depth)
		} else {
			// Reorg seems shallow enough to pull in all transactions into memory
			var discarded, included types.Transactions

			var (
				remove = tp.chain.GetBlock(oldHeader.Hash(), oldHeader.Number.Uint64())
				add    = tp.chain.GetBlock(newHeader.Hash(), newHeader.Number.Uint64())
			)

			for remove.NumberU64() > add.NumberU64() {
				discarded = append(discarded, remove.Txs...)
				if remove = tp.chain.GetBlock(remove.ParentHash(), remove.NumberU64()-1); remove == nil {
					log.Error("Unrooted old chain seen by tx pool", "block", oldHeader.Number, "hash", oldHeader.Hash())
					return
				}
			}

			for add.NumberU64() > remove.NumberU64() {
				included = append(included, add.Txs...)
				if add = tp.chain.GetBlock(add.ParentHash(), add.NumberU64()-1); add == nil {
					log.Error("Unrooted new chain seen by tx pool", "block", newHeader.Number, "hash", newHeader.Hash())
					return
				}
			}

			for remove.Hash() != add.Hash() {
				discarded = append(discarded, remove.Txs...)
				if remove = tp.chain.GetBlock(remove.ParentHash(), remove.NumberU64()-1); remove == nil {
					log.Error("Unrooted old chain seen by tx pool", "block", oldHeader.Number, "hash", oldHeader.Hash())
					return
				}
				included = append(included, add.Txs...)
				if add = tp.chain.GetBlock(add.ParentHash(), add.NumberU64()-1); add == nil {
					log.Error("Unrooted new chain seen by tx pool", "block", newHeader.Number, "hash", newHeader.Hash())
					return
				}
			}
			reinject = TxDifference(discarded, included)
		}

	}
	// Initialize the internal state to the current head
	if newHeader == nil {
		newHeader = tp.chain.CurrentBlock().Header() // Special case during testing
	}
	statedb, err := tp.chain.StateAt(newHeader.Root)
	if err != nil {
		log.Error("Failed to reset txpool state", "err", err)
		return
	}
	tp.currentState = statedb
	tp.currentMaxGas = newHeader.GasLimit

	// Inject any transactions discarded due to reorgs
	log.Debug("Reinjecting stale transactions", "count", len(reinject))
	senderCacher.recover(tp.signer, reinject)
	tp.addTxsLocked(reinject, false)

	tp.demoteUnexecutables()

	// todo Update all accounts to the latest known pending nonce
	// for addr, list := range pool.pending {
	// 	txs := list.Flatten() // Heavy but will be cached and is needed by the miner anyway
	// 	pool.pendingState.SetNonce(addr, txs[len(txs)-1].Nonce()+1)
	// }
	tp.promoteExecutables(nil)

}

// addTxsLocked attempts to queue a batch of transactions if they are valid,
// whilst assuming the transaction pool lock is already held.
func (tp *TxPool) addTxsLocked(txs []*types.Transaction, local bool) []error {
	// Add the batch of transaction, tracking the accepted ones
	dirty := make(map[common.Address]struct{})
	errs := make([]error, len(txs))

	for i, tx := range txs {
		var replace bool
		if replace, errs[i] = tp.add(tx, local); errs[i] == nil && !replace {
			from, _ := types.Sender(tp.signer, tx) // already validated
			dirty[from] = struct{}{}
		}
	}
	// Only reprocess the internal state if something was actually added
	if len(dirty) > 0 {
		addrs := make([]common.Address, 0, len(dirty))
		for addr := range dirty {
			addrs = append(addrs, addr)
		}
		tp.promoteExecutables(addrs)
	}
	return errs
}

// demoteUnexecutables removes invalid and processed transactions from the pools
// executable/pending queue and any subsequent transactions that become unexecutable
// are moved back into the future queue.
func (tp *TxPool) demoteUnexecutables() {
	// Iterate over all accounts and demote any non-executable transactions
	// for addr, list := range tp.pending {
	// 	nonce := tp.currentState.GetNonce(addr)

	// 	// Drop all transactions that are deemed too old (low nonce)
	// 	for _, tx := range list.Forward(nonce) {
	// 		hash := tx.Hash()
	// 		log.Trace("Removed old pending transaction", "hash", hash)
	// 		tp.all.Remove(hash)
	// 		tp.priced.Removed()
	// 	}
	// 	// Drop all transactions that are too costly (low balance or out of gas), and queue any invalids back for later
	// 	drops, invalids := list.Filter(tp.currentState.GetBalance(addr), tp.currentMaxGas)
	// 	for _, tx := range drops {
	// 		hash := tx.Hash()
	// 		log.Trace("Removed unpayable pending transaction", "hash", hash)
	// 		tp.all.Remove(hash)
	// 		tp.priced.Removed()
	// 	}
	// 	for _, tx := range invalids {
	// 		hash := tx.Hash()
	// 		log.Trace("Demoting pending transaction", "hash", hash)
	// 		tp.enqueueTx(hash, tx)
	// 	}
	// 	// If there's a gap in front, alert (should never happen) and postpone all transactions
	// 	if list.Len() > 0 && list.txs.Get(nonce) == nil {
	// 		for _, tx := range list.Cap(0) {
	// 			hash := tx.Hash()
	// 			log.Error("Demoting invalidated transaction", "hash", hash)
	// 			tp.enqueueTx(hash, tx)
	// 		}
	// 	}
	// 	// Delete the entire queue entry if it became empty.
	// 	if list.Empty() {
	// 		delete(tp.pending, addr)
	// 		delete(tp.beats, addr)
	// 	}
	// }
}

// promoteExecutables moves transactions that have become processable from the
// future queue to the set of pending transactions. During this process, all
// invalidated transactions (low nonce, low balance) are deleted.
func (tp *TxPool) promoteExecutables(accounts []common.Address) {
	// Track the promoted transactions to broadcast them at once
	// var promoted []*types.Transaction

	// Gather all the accounts potentially needing updates
	if accounts == nil {
		accounts = make([]common.Address, 0, len(tp.queue))
		for addr := range tp.queue {
			accounts = append(accounts, addr)
		}
	}
	// Iterate over all accounts and promote any executable transactions
	for _, addr := range accounts {
		list := tp.queue[addr]
		if list == nil {
			continue // Just in case someone calls with a non existing account
		}
		// todo add state check
		// // Drop all transactions that are deemed too old (low nonce)
		// for _, tx := range list.Forward(tp.currentState.GetNonce(addr)) {
		// 	hash := tx.Hash()
		// 	log.Trace("Removed old queued transaction", "hash", hash)
		// 	tp.all.Remove(hash)
		// 	tp.priced.Removed()
		// }
		// // Drop all transactions that are too costly (low balance or out of gas)
		// drops, _ := list.Filter(tp.currentState.GetBalance(addr), pool.currentMaxGas)
		// for _, tx := range drops {
		// 	hash := tx.Hash()
		// 	log.Trace("Removed unpayable queued transaction", "hash", hash)
		// 	tp.all.Remove(hash)
		// 	tp.priced.Removed()
		// }
		// // Gather all executable transactions and promote them
		// for _, tx := range list.Ready(tp.pendingState.GetNonce(addr)) {
		// 	hash := tx.Hash()
		// 	if tp.promoteTx(addr, hash, tx) {
		// 		log.Trace("Promoting queued transaction", "hash", hash)
		// 		promoted = append(promoted, tx)
		// 	}
		// }
		// Drop all transactions over the allowed limit
		if !tp.locals.contains(addr) {
			for _, tx := range list.Cap(int(tp.config.AccountQueue)) {
				hash := tx.Hash()
				tp.all.Remove(hash)
				tp.priced.Removed()
				log.Trace("Removed cap-exceeding queued transaction", "hash", hash)
			}
		}
		// Delete the entire queue entry if it became empty.
		if list.Empty() {
			delete(tp.queue, addr)
		}
	}
	// Notify subsystem for new promoted transactions.
	// if len(promoted) > 0 {
	// 	go tp.txFeed.Send(NewTxsEvent{promoted})
	// }
	// If the pending limit is overflown, start equalizing allowances
	pending := uint64(0)
	for _, list := range tp.pending {
		pending += uint64(list.Len())
	}
	if pending > tp.config.GlobalSlots {
		// Assemble a spam order to penalize large transactors first
		spammers := prque.New()
		for addr, list := range tp.pending {
			// Only evict transactions from high rollers
			if !tp.locals.contains(addr) && uint64(list.Len()) > tp.config.AccountSlots {
				spammers.Push(addr, float32(list.Len()))
			}
		}
		// Gradually drop transactions from offenders
		offenders := []common.Address{}
		for pending > tp.config.GlobalSlots && !spammers.Empty() {
			// Retrieve the next offender if not local address
			offender, _ := spammers.Pop()
			offenders = append(offenders, offender.(common.Address))

			// Equalize balances until all the same or below threshold
			if len(offenders) > 1 {
				// Calculate the equalization threshold for all current offenders
				threshold := tp.pending[offender.(common.Address)].Len()

				// Iteratively reduce all offenders until below limit or threshold reached
				for pending > tp.config.GlobalSlots && tp.pending[offenders[len(offenders)-2]].Len() > threshold {
					for i := 0; i < len(offenders)-1; i++ {
						list := tp.pending[offenders[i]]
						for _, tx := range list.Cap(list.Len() - 1) {
							// Drop the transaction from the global pools too
							hash := tx.Hash()
							tp.all.Remove(hash)
							tp.priced.Removed()

							// todo Update the account nonce to the dropped transaction
							// if nonce := tx.Nonce(); tp.pendingState.GetNonce(offenders[i]) > nonce {
							// 	tp.pendingState.SetNonce(offenders[i], nonce)
							// }
							log.Trace("Removed fairness-exceeding pending transaction", "hash", hash)
						}
						pending--
					}
				}
			}
		}
		// If still above threshold, reduce to limit or min allowance
		if pending > tp.config.GlobalSlots && len(offenders) > 0 {
			for pending > tp.config.GlobalSlots && uint64(tp.pending[offenders[len(offenders)-1]].Len()) > tp.config.AccountSlots {
				for _, addr := range offenders {
					list := tp.pending[addr]
					for _, tx := range list.Cap(list.Len() - 1) {
						// Drop the transaction from the global pools too
						hash := tx.Hash()
						tp.all.Remove(hash)
						tp.priced.Removed()

						// Update the account nonce to the dropped transaction
						// if nonce := tx.Nonce(); pool.pendingState.GetNonce(addr) > nonce {
						// 	pool.pendingState.SetNonce(addr, nonce)
						// }
						log.Trace("Removed fairness-exceeding pending transaction", "hash", hash)
					}
					pending--
				}
			}
		}
	}
	// If we've queued more transactions than the hard limit, drop oldest ones
	queued := uint64(0)
	for _, list := range tp.queue {
		queued += uint64(list.Len())
	}
	if queued > tp.config.GlobalQueue {
		// Sort all accounts with queued transactions by heartbeat
		addresses := make(addresssByHeartbeat, 0, len(tp.queue))
		for addr := range tp.queue {
			if !tp.locals.contains(addr) { // don't drop locals
				addresses = append(addresses, addressByHeartbeat{addr, tp.beats[addr]})
			}
		}
		sort.Sort(addresses)

		// Drop transactions until the total is below the limit or only locals remain
		for drop := queued - tp.config.GlobalQueue; drop > 0 && len(addresses) > 0; {
			addr := addresses[len(addresses)-1]
			list := tp.queue[addr.address]

			addresses = addresses[:len(addresses)-1]

			// Drop all transactions if they are less than the overflow
			if size := uint64(list.Len()); size <= drop {
				for _, tx := range list.Flatten() {
					tp.removeTx(tx.Hash(), true)
				}
				drop -= size
				continue
			}
			// Otherwise drop only last few transactions
			txs := list.Flatten()
			for i := len(txs) - 1; i >= 0 && drop > 0; i-- {
				tp.removeTx(txs[i].Hash(), true)
				drop--
			}
		}
	}
}

// add validates a transaction and inserts it into the non-executable queue for
// later pending promotion and execution. If the transaction is a replacement for
// an already pending or queued one, it overwrites the previous and returns this
// so outer code doesn't uselessly call promote.
//
// If a newly added transaction is marked as local, its sending account will be
// whitelisted, preventing any associated transaction from being dropped out of
// the pool due to pricing constraints.
func (tp *TxPool) add(tx *types.Transaction, local bool) (bool, error) {
	//If the transaction is already known, discard it
	hash := tx.Hash()
	if tp.all.Get(hash) != nil {
		log.Trace("Discarding already known transaction", "hash", hash)
		return false, fmt.Errorf("known transaction: %x", hash)
	}
	// If the transaction fails basic validation, discard it
	if err := tp.validateTx(tx, local); err != nil {
		log.Trace("Discarding invalid transaction", "hash", hash, "err", err)
		return false, err
	}
	// If the transaction pool is full, discard underpriced transactions
	if uint64(tp.all.Count()) >= tp.config.GlobalSlots+tp.config.GlobalQueue {
		// If the new transaction is underpriced, don't accept it
		if !local && tp.priced.Underpriced(tx, tp.locals) {
			log.Trace("Discarding underpriced transaction", "hash", hash, "price", tx.GasPrice())
			return false, ErrUnderpriced
		}
		// New transaction is better than our worse ones, make room for it
		drop := tp.priced.Discard(tp.all.Count()-int(tp.config.GlobalSlots+tp.config.GlobalQueue-1), tp.locals)
		for _, tx := range drop {
			log.Trace("Discarding freshly underpriced transaction", "hash", tx.Hash(), "price", tx.GasPrice())
			tp.removeTx(tx.Hash(), false)
		}
	}
	// If the transaction is replacing an already pending one, do directly
	from, _ := types.Sender(tp.signer, tx) // already validated
	if list := tp.pending[from]; list != nil && list.Overlaps(tx) {
		// Nonce already pending, check if required price bump is met
		inserted, old := list.Add(tx, tp.config.PriceBump)
		if !inserted {
			return false, ErrReplaceUnderpriced
		}
		// New transaction is better, replace old one
		if old != nil {
			tp.all.Remove(old.Hash())
			tp.priced.Removed()
		}
		tp.all.Add(tx)
		tp.priced.Put(tx)
		tp.journalTx(from, tx)

		log.Trace("Pooled new executable transaction", "hash", hash, "from", from, "to", tx.To())

		// We've directly injected a replacement transaction, notify subsystems
		// go tp.txFeed.Send(NewTxsEvent{types.Transactions{tx}})

		return old != nil, nil
	}
	// New transaction isn't replacing a pending one, push into queue
	replace, err := tp.enqueueTx(hash, tx)
	if err != nil {
		return false, err
	}
	// Mark local addresses and journal local transactions
	if local {
		tp.locals.add(from)
	}
	tp.journalTx(from, tx)

	log.Trace("Pooled new future transaction", "hash", hash, "from", from, "to", tx.To())
	return replace, nil
}

// enqueueTx inserts a new transaction into the non-executable transaction queue.
//
// Note, this method assumes the pool lock is held!
func (tp *TxPool) enqueueTx(hash common.Hash, tx *types.Transaction) (bool, error) {
	// Try to insert the transaction into the future queue
	from, _ := types.Sender(tp.signer, tx) // already validated
	if tp.queue[from] == nil {
		tp.queue[from] = newTxList(false)
	}
	inserted, old := tp.queue[from].Add(tx, tp.config.PriceBump)
	if !inserted {
		// An older transaction was better, discard this
		return false, ErrReplaceUnderpriced
	}
	// Discard any previous transaction and mark this
	if old != nil {
		tp.all.Remove(old.Hash())
		tp.priced.Removed()
	}
	if tp.all.Get(hash) == nil {
		tp.all.Add(tx)
		tp.priced.Put(tx)
	}
	return old != nil, nil
}

// removeTx removes a single transaction from the queue, moving all subsequent
// transactions back to the future queue.
func (tp *TxPool) removeTx(hash common.Hash, outofbound bool) {
	// Fetch the transaction we wish to delete
	tx := tp.all.Get(hash)
	if tx == nil {
		return
	}

	addr, _ := types.Sender(tp.signer, tx) // already validated during insertion

	// Remove it from the list of known transactions
	tp.all.Remove(hash)
	if outofbound {
		tp.priced.Removed()
	}
	// Remove the transaction from the pending lists and reset the account nonce
	if pending := tp.pending[addr]; pending != nil {
		if removed, invalids := pending.Remove(tx); removed {
			// If no more pending transactions are left, remove the list
			if pending.Empty() {
				delete(tp.pending, addr)
				delete(tp.beats, addr)
			}
			// Postpone any invalidated transactions
			for _, tx := range invalids {
				tp.enqueueTx(tx.Hash(), tx)
			}
			// todo Update the account nonce if needed
			// if nonce := tx.Nonce(); tp.pendingState.GetNonce(addr) > nonce {
			// 	tp.pendingState.SetNonce(addr, nonce)
			// }
			return
		}
	}
	// Transaction is in the future queue
	if future := tp.queue[addr]; future != nil {
		future.Remove(tx)
		if future.Empty() {
			delete(tp.queue, addr)
		}
	}
}

// journalTx adds the specified transaction to the local disk journal if it is
// deemed to have been sent from a local account.
func (tp *TxPool) journalTx(from common.Address, tx *types.Transaction) {
	// Only journal if it's enabled and the transaction is local
	if tp.journal == nil || !tp.locals.contains(from) {
		return
	}
	if err := tp.journal.insert(tx); err != nil {
		log.Warn("Failed to journal local transaction", "err", err)
	}
}

// validateTx checks whether a transaction is valid according to the consensus
// rules and adheres to some heuristic limits of the local node (price and size).
func (tp *TxPool) validateTx(tx *types.Transaction, local bool) error {
	// Heuristic limit, reject transactions over 32KB to prevent DOS attacks
	if tx.Size() > 32*1024 {
		return ErrOversizedData
	}

	// todo check  transactions value not be negative.

	// Ensure the transaction doesn't exceed the current block limit gas.
	if tp.currentMaxGas < tx.Gas() {
		return ErrGasLimit
	}
	// Make sure the transaction is signed properly
	from, err := types.Sender(tp.signer, tx)
	if err != nil {
		return ErrInvalidSender
	}
	// Drop non-local transactions under our own minimal accepted gas price
	local = local || tp.locals.contains(from) // account may be local even if the transaction arrived from the network
	if !local && tp.gasPrice.Cmp(tx.GasPrice()) > 0 {
		return ErrUnderpriced
	}

	// todo check state nonce and value
	intrGas, err := IntrinsicGas(tx.Payload(), tx.To() == nil)
	if err != nil {
		return err
	}
	if tx.Gas() < intrGas {
		return ErrIntrinsicGas
	}
	return nil
}

// AddLocal enqueues a single transaction into the pool if it is valid, marking
// the sender as a local one in the mean time, ensuring it goes around the local
// pricing constraints.
func (tp *TxPool) AddLocal(tx *types.Transaction) error {
	return tp.addTx(tx, !tp.config.NoLocals)
}

// AddLocals enqueues a batch of transactions into the pool if they are valid,
// marking the senders as a local ones in the mean time, ensuring they go around
// the local pricing constraints.
func (tp *TxPool) AddLocals(txs []*types.Transaction) []error {
	return tp.addTxs(txs, !tp.config.NoLocals)
}

// AddRemote enqueues a single transaction into the pool if it is valid. If the
// sender is not among the locally tracked ones, full pricing constraints will
// apply.
func (tp *TxPool) AddRemote(tx *types.Transaction) error {
	return tp.addTx(tx, false)
}

// AddRemotes enqueues a batch of transactions into the pool if they are valid.
// If the senders are not among the locally tracked ones, full pricing constraints
// will apply.
func (tp *TxPool) AddRemotes(txs []*types.Transaction) []error {
	return tp.addTxs(txs, false)
}

// addTx enqueues a single transaction into the pool if it is valid.
func (tp *TxPool) addTx(tx *types.Transaction, local bool) error {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	// Try to inject the transaction and update any state
	replace, err := tp.add(tx, local)
	if err != nil {
		return err
	}
	// If we added a new transaction, run promotion checks and return
	if !replace {
		from, _ := types.Sender(tp.signer, tx) // already validated
		tp.promoteExecutables([]common.Address{from})
	}
	return nil
}

// addTxs attempts to queue a batch of transactions if they are valid.
func (tp *TxPool) addTxs(txs []*types.Transaction, local bool) []error {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	return tp.addTxsLocked(txs, local)
}

// local retrieves all currently known local transactions, groupped by origin
// account and sorted by nonce. The returned transaction set is a copy and can be
// freely modified by calling code.
func (tp *TxPool) local() map[common.Address]types.Transactions {
	txs := make(map[common.Address]types.Transactions)
	for addr := range tp.locals.accounts {
		if pending := tp.pending[addr]; pending != nil {
			txs[addr] = append(txs[addr], pending.Flatten()...)
		}
		if queued := tp.queue[addr]; queued != nil {
			txs[addr] = append(txs[addr], queued.Flatten()...)
		}
	}
	return txs
}
