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
package types

import (
	"math/big"
	"testing"

	"github.com/zipper-project/z0/common"
)

var (
	testTx = NewTransaction(
		3,
		"123456",
		2000,
		big.NewInt(1),
		[]byte("test transaction"),
	)
)

func TestTransactionEncodeAndDecode(t *testing.T) {
	bytes, _ := testTx.MarshalJSON()
	t.Log(string(bytes))

	testTx.WithInput([]In{amInput})
	testTx.WithOutput([]Out{amOutput})
	{
		bytes, _ := testTx.EncodeRLP()
		newTx := &Transaction{}
		newTx.DecodeRLP(bytes)

		common.AssertEquals(t, newTx.Data(), testTx.Data())
		common.AssertEquals(t, newTx.Gas(), testTx.Gas())
		common.AssertEquals(t, newTx.GasPrice(), testTx.GasPrice())
		common.AssertEquals(t, newTx.Nonce(), testTx.Nonce())
		common.AssertEquals(t, newTx.AssertID(), testTx.AssertID())

		tmpBytes, _ := newTx.EncodeRLP()
		common.AssertEquals(t, bytes, tmpBytes)

	}
}
