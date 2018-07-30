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
	"io/ioutil"
	"math/big"
	"os"
	"testing"

	"github.com/zipper-project/z0/common"
	"github.com/zipper-project/z0/types"
)

func TestTxJournal(t *testing.T) {
	// Create a temporary file for the journal
	file, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatalf("failed to create temporary journal: %v", err)
	}
	txj := newTxJournal(file.Name())
	defer os.Remove(file.Name())

	txsMap := make(map[common.Address]types.Transactions)
	tx := types.NewTransaction(1, 0, big.NewInt(200), nil)
	txsMap[common.Address{}] = types.Transactions{
		tx,
	}
	if err := txj.rotate(txsMap); err != nil {
		t.Fatalf("Failed to rotate transaction journal: %v", err)
	}

	if err := txj.close(); err != nil {
		t.Fatalf("Failed to close transaction journal: %v", err)
	}

	txjLoad := newTxJournal(file.Name())

	if err := txjLoad.load(func(txs []*types.Transaction) []error {
		common.AssertEquals(t, tx.Nonce(), txs[0].Nonce())
		return nil
	}); err != nil {
		t.Fatalf("Failed to close transaction journal: %v", err)
	}
}
