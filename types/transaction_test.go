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
	addr     = common.HexToAddress("0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed")
	assertID = common.HexToAddress("0xcc8507ed53e44c9d86a158e876c151634247f514")
	amInput  = AMInput{AssertID: assertID, Payload: []byte("payload")}
	amOutput = AMOutput{AssertID: assertID, Address: addr, Value: big.NewInt(10000)}

	testTx = NewTransaction(
		3,
		2000,
		big.NewInt(1),
		[]byte("test transaction"),
	)
)

func TestTransactionEncodeAndDecode(t *testing.T) {
	testTx.WithInput(amInput)
	testTx.WithOutput(amOutput)
	bytes, err := testTx.MarshalJSON()
	t.Log(string(bytes), err)

	{
		bytes, _ := testTx.EncodeRLP()
		newTx := &Transaction{}
		newTx.DecodeRLP(bytes)

		common.AssertEquals(t, newTx.Extra(), testTx.Extra())
		common.AssertEquals(t, newTx.Gas(), testTx.Gas())
		common.AssertEquals(t, newTx.GasPrice(), testTx.GasPrice())
		common.AssertEquals(t, newTx.Nonce(), testTx.Nonce())
		common.AssertEquals(t, newTx.GetInputs(), []interface{}{amInput})
		common.AssertEquals(t, newTx.GetOutputs(), []interface{}{amOutput})

		tmpBytes, _ := newTx.EncodeRLP()
		common.AssertEquals(t, bytes, tmpBytes)

	}
}
