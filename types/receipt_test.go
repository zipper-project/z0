// Copyright 2018 The zipper Authors
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
	"testing"

	"github.com/zipper-project/z0/common"
)

var (
	testR = NewReceipt([]byte("root"), false, 1000)
)

func TestReceiptEncodeAndDecode(t *testing.T) {
	bytes, _ := testR.EncodeRLP()
	newR := &Receipt{}
	newR.DecodeRLP(bytes)
	common.AssertEquals(t, testR.PostState, newR.PostState)
	common.AssertEquals(t, testR.Status, newR.Status)
	common.AssertEquals(t, testR.CumulativeGasUsed, newR.CumulativeGasUsed)

	tmpBytes, _ := newR.EncodeRLP()
	common.AssertEquals(t, bytes, tmpBytes)

}
