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
	"math/big"
	"testing"

	"github.com/zipper-project/z0/types"
)

func TestTxPriceList(t *testing.T) {
	txlk := newTxLookup()
	txpl := newTxPricedList(txlk)
	txpl.Put(types.NewTransaction(2, "", 0, big.NewInt(200), nil))
	txpl.Put(types.NewTransaction(1, "", 0, big.NewInt(200), nil))
	txpl.Put(types.NewTransaction(4, "", 0, big.NewInt(400), nil))
	txpl.Put(types.NewTransaction(3, "", 0, big.NewInt(100), nil))

	// test Cap

	// test Underpriced

	// test Discard

}
