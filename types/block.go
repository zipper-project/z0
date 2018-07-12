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
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math/big"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/zipper-project/z0/common"
	"github.com/zipper-project/z0/utils/rlp"
)

// A BlockNonce is a 64-bit hash which proves (combined with the
// mix-hash) that a sufficient amount of computation has been carried
// out on a block.
type BlockNonce [8]byte

// EncodeNonce converts the given integer to a block nonce.
func EncodeNonce(i uint64) BlockNonce {
	var n BlockNonce
	binary.BigEndian.PutUint64(n[:], i)
	return n
}

// Uint64 returns the integer value of a block nonce.
func (n BlockNonce) Uint64() uint64 {
	return binary.BigEndian.Uint64(n[:])
}

// MarshalText encodes n as a hex string with 0x prefix.
func (n BlockNonce) MarshalText() ([]byte, error) {
	return hexutil.Bytes(n[:]).MarshalText()
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (n *BlockNonce) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("BlockNonce", input, n[:])
}

// Header represents a block header in the blockchain.
type Header struct {
	ParentHash  common.Hash    `json:"parentHash"      `
	Coinbase    common.Address `json:"miner"           `
	Root        common.Hash    `json:"stateRoot"       `
	TxHash      common.Hash    `json:"transactionsRoot"`
	ReceiptHash common.Hash    `json:"receiptsRoot"    `
	Difficulty  *big.Int       `json:"difficulty"      `
	Number      *big.Int       `json:"number"          `
	GasLimit    uint64         `json:"gasLimit"        `
	GasUsed     uint64         `json:"gasUsed"         `
	Time        *big.Int       `json:"timestamp"       `
	Extra       []byte         `json:"extraData"       `
	MixDigest   common.Hash    `json:"mixHash"         `
	Nonce       BlockNonce     `json:"nonce"           `
}

// Hash returns the block hash of the header, which is simply the keccak256 hash of its
// RLP encoding.
func (h *Header) Hash() common.Hash {
	return rlpHash(h)
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}

// EncodeRLP serializes b into the  RLP block header format.
func (h *Header) EncodeRLP() ([]byte, error) { return rlp.EncodeToBytes(h) }

// DecodeRLP decodes the header
func (h *Header) DecodeRLP(input []byte) error { return rlp.Decode(bytes.NewReader(input), &h) }

// Marshal encodes the web3 RPC block header format.
func (h *Header) Marshal() ([]byte, error) { return json.Marshal(h) }

// Unmarshal decodes the web3 RPC block header format.
func (h *Header) Unmarshal(input []byte) error { return json.Unmarshal(input, h) }

// Block represents an entire block in the blockchain.
type Block struct {
	Head *Header
	Txs  Transactions

	// caches
	hash atomic.Value
	size atomic.Value

	// Td is used by package core to store the total difficulty
	// of the chain up to and including the block.
	td *big.Int

	// These fields are used by package eth to track
	// inter-peer block relay.
	receivedAt   time.Time
	receivedFrom interface{}
}

func (b *Block) Number() *big.Int     { return new(big.Int).Set(b.Head.Number) }
func (b *Block) GasLimit() uint64     { return b.Head.GasLimit }
func (b *Block) GasUsed() uint64      { return b.Head.GasUsed }
func (b *Block) Difficulty() *big.Int { return new(big.Int).Set(b.Head.Difficulty) }
func (b *Block) Time() *big.Int       { return new(big.Int).Set(b.Head.Time) }

func (b *Block) NumberU64() uint64      { return b.Head.Number.Uint64() }
func (b *Block) MixDigest() common.Hash { return b.Head.MixDigest }

func (b *Block) Nonce() uint64            { return binary.BigEndian.Uint64(b.Head.Nonce[:]) }
func (b *Block) Coinbase() common.Address { return b.Head.Coinbase }
func (b *Block) Root() common.Hash        { return b.Head.Root }
func (b *Block) ParentHash() common.Hash  { return b.Head.ParentHash }
func (b *Block) TxHash() common.Hash      { return b.Head.TxHash }
func (b *Block) ReceiptHash() common.Hash { return b.Head.ReceiptHash }
func (b *Block) Extra() []byte            { return common.CopyBytes(b.Head.Extra) }
func (b *Block) Header() *Header          { return CopyHeader(b.Head) }

// EncodeRLP serializes b into the RLP block format.
func (b *Block) EncodeRLP() ([]byte, error) {
	for _, tx := range b.Txs {
		tx.data.Inputs, tx.data.Outputs = serialize(tx.inputs, tx.outputs, false)
	}
	return rlp.EncodeToBytes(b)
}

// DecodeRLP decodes the block
func (b *Block) DecodeRLP(input []byte) error {
	err := rlp.Decode(bytes.NewReader(input), &b)
	if err == nil {
		b.size.Store(common.StorageSize(len(input)))
		for _, tx := range b.Txs {
			tx.inputs, tx.outputs = deserialize(tx.data.Inputs, tx.data.Outputs)
		}
	}
	return err
}

// Marshal encodes the web3 RPC block format.
func (b *Block) Marshal() ([]byte, error) {
	type Block struct {
		Header       *Header
		Transactions Transactions
	}
	for _, tx := range b.Txs {
		tx.data.Inputs, tx.data.Outputs = serialize(tx.inputs, tx.outputs, true)
	}
	var block Block
	block.Header = b.Head
	block.Transactions = b.Txs
	return json.Marshal(block)
}

// CopyHeader creates a deep copy of a block header to prevent side effects from
// modifying a header variable.
func CopyHeader(h *Header) *Header {
	cpy := *h
	if cpy.Time = new(big.Int); h.Time != nil {
		cpy.Time.Set(h.Time)
	}
	if cpy.Difficulty = new(big.Int); h.Difficulty != nil {
		cpy.Difficulty.Set(h.Difficulty)
	}
	if cpy.Number = new(big.Int); h.Number != nil {
		cpy.Number.Set(h.Number)
	}
	if len(h.Extra) > 0 {
		cpy.Extra = make([]byte, len(h.Extra))
		copy(cpy.Extra, h.Extra)
	}
	return &cpy
}
