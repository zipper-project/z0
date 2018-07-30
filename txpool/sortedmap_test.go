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
	"testing"

	"github.com/zipper-project/z0/common"

	"github.com/zipper-project/z0/types"
)

func TestPutAndGetAndRemove(t *testing.T) {
	sm := newTxSortedMap()
	nonce := uint64(2)
	tx := types.NewTransaction(nonce, 0, nil, nil)
	sm.Put(tx)
	common.AssertEquals(t, sm.Get(nonce), tx)

	// test remove
	common.AssertEquals(t, sm.Remove(nonce), true)
	common.AssertEquals(t, sm.Remove(nonce), false)
}

//func TestSortedMap(t *testing.T) {
//	sm := newTxSortedMap()
//	txs := []*types.Transaction{
//		types.NewTransaction(4, "", 0, nil, nil),
//		types.NewTransaction(1, "", 0, nil, nil),
//		types.NewTransaction(2, "", 0, nil, nil),
//		types.NewTransaction(3, "", 0, nil, nil),
//	}
//	for _, v := range txs {
//		sm.Put(v)
//	}
//
//	// test cap
//	sm.Cap(3)
//
//	common.AssertEquals(t, sm.Len(), 3)
//
//	// test ready
//	readyTxs := sm.Ready(0)
//	sort1 := []*types.Transaction{
//		types.NewTransaction(1, "", 0, nil, nil),
//		types.NewTransaction(2, "", 0, nil, nil),
//		types.NewTransaction(3, "", 0, nil, nil),
//	}
//
//	for k, v := range readyTxs {
//		common.AssertEquals(t, v.Nonce(), sort1[k].Nonce())
//	}
//
//	for _, v := range txs {
//		sm.Put(v)
//	}
//
//	// test  flatten
//	flatten := sm.Flatten()
//	for k, v := range flatten {
//		common.AssertEquals(t, v.Nonce(), txs[k].Nonce())
//	}
//
//	// test Filter
//	sm.Filter(func(tx *types.Transaction) bool {
//		for _, v := range txs {
//			if tx.Nonce() == v.Nonce() {
//				return true
//			}
//		}
//		return false
//	})
//	common.AssertEquals(t, sm.Len(), 0)
//
//}
