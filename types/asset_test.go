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

// Package types contains data types related to Z0.
package types

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/zipper-project/z0/common"
)

var (
	addr     = common.HexToAddress("0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed")
	amInput  = &AMInput{Amount: big.NewInt(1000)}
	amOutput = &AMOutput{Address: &addr}
	in       In
	out      Out
)

func TestEncodeRLPAndDecodeRLP(t *testing.T) {
	{
		in = amInput
		bytes := in.EncodeRLP()
		tmpIn := InDecodeRLP(bytes)
		tmpBytes := tmpIn.EncodeRLP()
		common.AssertEquals(t, bytes, tmpBytes)

		out = amOutput
		bytes = out.EncodeRLP()
		tmpOut := OutDecodeRLP(bytes)
		tmpBytes = tmpOut.EncodeRLP()
		common.AssertEquals(t, bytes, tmpBytes)
	}
}

func TestMarshalJSONAndUnMarshalJSON(t *testing.T) {
	{
		in = amInput
		tmpAmInput := &AMInput{}
		bytes := in.MarshalJSON()
		json.Unmarshal(bytes, tmpAmInput)
		tmpBytes, _ := json.Marshal(tmpAmInput)
		common.AssertEquals(t, bytes, tmpBytes)

		out = amOutput
		tmpOutput := &AMOutput{}
		bytes = out.MarshalJSON()
		json.Unmarshal(bytes, tmpOutput)
		tmpBytes, _ = json.Marshal(tmpOutput)
		common.AssertEquals(t, bytes, tmpBytes)
	}
}
