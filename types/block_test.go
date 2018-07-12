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
	th = &Header{
		ParentHash: common.HexToHash("0a5843ac1cb04865017cb35a57b50b07084e5fcee39b5acadade33149f4fff9e"),
		Coinbase:   common.HexToAddress("8888f1f195afa192cfee860698584c030f4c9db1"),
		Root:       common.HexToHash("ef1552a40b7165c3cd773806b9e0c165b75356e0314bf0706f279c729f51e017"),
		TxHash:     common.HexToHash("0a5843ac1cb04865017cb35a57b50b07084e5fcee39b5acadade33149f4fff9e"),
		Difficulty: big.NewInt(131072),
		Number:     big.NewInt(100),
		GasLimit:   uint64(3141592),
		GasUsed:    uint64(21000),
		Time:       big.NewInt(1426516743),
		Extra:      []byte("test Header"),
		MixDigest:  common.HexToHash("bd4472abb6659ebe3ee06ee4d7b72a00a9f4d001caca51342001075469aff498"),
		Nonce:      EncodeNonce(uint64(0xa13a5a8c8f2bb1c4)),
	}
	b = &Block{
		Head: th,
		Txs:  []*Transaction{testTx},
	}
)

func TestBlockEncodeRLPAndDecodeRLP(t *testing.T) {
	bytes, _ := b.Marshal()
	t.Log(string(bytes))

	bytes, _ = b.EncodeRLP()
	newBlock := &Block{}
	newBlock.DecodeRLP(bytes)
	tmpBytes, _ := newBlock.EncodeRLP()
	common.AssertEquals(t, bytes, tmpBytes)
}

func TestBlockHeaderMarshalAndUnmarshal(t *testing.T) {
	bytes, _ := th.Marshal()
	newHeader := &Header{}
	newHeader.Unmarshal(bytes)
	tmpBytes, _ := newHeader.Marshal()
	common.AssertEquals(t, bytes, tmpBytes)
}

func TestBlockHeaderEncodeRLPAndDecodeRLP(t *testing.T) {
	bytes, _ := th.EncodeRLP()
	newHeader := &Header{}
	newHeader.DecodeRLP(bytes)
	tmpBytes, _ := newHeader.EncodeRLP()
	common.AssertEquals(t, bytes, tmpBytes)
}
