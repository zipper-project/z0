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
	"bytes"
	"encoding/json"
	"errors"
	"math/big"
	"sync"
	"sync/atomic"

	"github.com/zipper-project/z0/common"
	"github.com/zipper-project/z0/crypto"
	"github.com/zipper-project/z0/utils/rlp"
)

// ErrInvalidSig invalid signature
var ErrInvalidSig = errors.New("invalid transaction v, r, s values")

// Transaction represents an entire transaction in the block.
type Transaction struct {
	Data    txdata
	Inputs  []In
	Outputs []Out

	// caches
	hash atomic.Value
	size atomic.Value
	from atomic.Value
}

type txdata struct {
	AccountNonce uint64   `json:"nonce"   `
	Price        *big.Int `json:"gasPrice"`
	GasLimit     uint64   `json:"gas"     `
	AssertID     string   `json:"assertid"`
	Inputs       [][]byte `json:"inputs"`
	Outputs      [][]byte `json:"outputs"`
	Payload      []byte   `json:"input"   `

	// Signature values
	V *big.Int `json:"v"`
	R *big.Int `json:"r"`
	S *big.Int `json:"s"`

	// This is only used when marshaling to JSON.
	Hash *common.Hash `json:"hash" rlp:"-"`
}

// NewTransaction initialize transaction
func NewTransaction(nonce uint64, assertID string, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction {
	return newTransaction(nonce, assertID, gasLimit, gasPrice, data)
}

func newTransaction(nonce uint64, assertID string, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction {
	if len(data) > 0 {
		data = common.CopyBytes(data)
	}
	d := txdata{
		AccountNonce: nonce,
		Payload:      data,
		GasLimit:     gasLimit,
		AssertID:     assertID,
		Price:        new(big.Int),
		V:            new(big.Int),
		R:            new(big.Int),
		S:            new(big.Int),
	}
	if gasPrice != nil {
		d.Price.Set(gasPrice)
	}
	return &Transaction{Data: d}
}

// WithInput add transaction input
func (tx *Transaction) WithInput(inputs []In) { tx.Inputs = inputs }

// WithOutput add transaction output
func (tx *Transaction) WithOutput(outputs []Out) { tx.Outputs = outputs }

// EncodeRLP implements rlp.Encoder
func (tx *Transaction) EncodeRLP() ([]byte, error) {
	tx.Data.Inputs, tx.Data.Outputs = serialize(tx.Inputs, tx.Outputs, false)
	return rlp.EncodeToBytes(&tx.Data)
}

// DecodeRLP implements rlp.Decoder
func (tx *Transaction) DecodeRLP(data []byte) error {
	err := rlp.Decode(bytes.NewReader(data), &tx.Data)
	if err == nil {
		tx.size.Store(common.StorageSize(len(data)))
		tx.Inputs, tx.Outputs = deserialize(tx.Data.Inputs, tx.Data.Outputs)
	}
	return err
}

// MarshalJSON encodes the web3 RPC transaction format.
func (tx *Transaction) MarshalJSON() ([]byte, error) {
	hash := tx.Hash()
	data := tx.Data
	data.Hash = &hash
	tx.Data.Inputs, tx.Data.Outputs = serialize(tx.Inputs, tx.Outputs, true)
	return json.Marshal(&data)
}

// UnmarshalJSON decodes the web3 RPC transaction format.
func (tx *Transaction) UnmarshalJSON(input []byte) error {
	var dec txdata
	if err := json.Unmarshal(input, dec); err != nil {
		return err
	}
	var V byte
	if isProtectedV(dec.V) {
		chainID := deriveChainID(dec.V).Uint64()
		V = byte(dec.V.Uint64() - 35 - 2*chainID)
	} else {
		V = byte(dec.V.Uint64() - 27)
	}
	if !crypto.ValidateSignatureValues(V, dec.R, dec.S, false) {
		return ErrInvalidSig
	}

	*tx = Transaction{Data: dec}
	return nil
}

// Hash hashes the RLP encoding of tx.
// It uniquely identifies the transaction.
func (tx *Transaction) Hash() common.Hash {
	if hash := tx.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}
	v := rlpHash(tx)
	tx.hash.Store(v)
	return v
}

func (tx *Transaction) Payload() []byte       { return common.CopyBytes(tx.Data.Payload) }
func (tx *Transaction) Gas() uint64        { return tx.Data.GasLimit }
func (tx *Transaction) GasPrice() *big.Int { return new(big.Int).Set(tx.Data.Price) }
func (tx *Transaction) Nonce() uint64      { return tx.Data.AccountNonce }
func (tx *Transaction) AssertID() string   { return tx.Data.AssertID }

// To returns the recipient address of the transaction.
// It returns nil if the transaction is a contract creation.
func (tx *Transaction) To() *common.Address {
	//todo
	return nil
}

// Size returns the true RLP encoded storage size of the transaction, either by
// encoding and returning it, or returning a previsouly cached value.
func (tx *Transaction) Size() common.StorageSize {
	if size := tx.size.Load(); size != nil {
		return size.(common.StorageSize)
	}
	bytes, _ := tx.EncodeRLP()

	tx.size.Store(common.StorageSize(len(bytes)))
	return common.StorageSize(len(bytes))
}

// ChainID returns which chain id this transaction was signed for (if at all)
func (tx *Transaction) ChainID() *big.Int {
	return deriveChainID(tx.Data.V)
}

// Protected returns whether the transaction is protected from replay protection.
func (tx *Transaction) Protected() bool {
	return isProtectedV(tx.Data.V)
}

// WithSignature returns a new transaction with the given signature.
// This signature needs to be formatted as described in the yellow paper (v+27).
func (tx *Transaction) WithSignature(signer Signer, sig []byte) (*Transaction, error) {
	r, s, v, err := signer.SignatureValues(tx, sig)
	if err != nil {
		return nil, err
	}
	cpy := &Transaction{Data: tx.Data}
	cpy.Data.R, cpy.Data.S, cpy.Data.V = r, s, v
	return cpy, nil
}

func isProtectedV(V *big.Int) bool {
	if V.BitLen() <= 8 {
		v := V.Uint64()
		return v != 27 && v != 28
	}
	// anything not 27 or 28 are considered unprotected
	return true
}

// deriveChainId derives the chain id from the given v parameter
func deriveChainID(v *big.Int) *big.Int {
	if v.BitLen() <= 64 {
		v := v.Uint64()
		if v == 27 || v == 28 {
			return new(big.Int)
		}
		return new(big.Int).SetUint64((v - 35) / 2)
	}
	v = new(big.Int).Sub(v, big.NewInt(35))
	return v.Div(v, big.NewInt(2))
}

// serialize if flag = true ,marshal to json else encodeRLP
func serialize(inputs []In, outputs []Out, flag bool) (ipdata, opdata [][]byte) {
	ipdata, opdata = make([][]byte, len(inputs)), make([][]byte, len(outputs))
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		for k, in := range inputs {
			if flag {
				ipdata[k] = in.MarshalJSON()
			} else {
				ipdata[k] = in.EncodeRLP()
			}
		}
		wg.Done()
	}()
	go func() {
		for k, out := range outputs {
			if flag {
				opdata[k] = out.MarshalJSON()
			} else {
				opdata[k] = out.EncodeRLP()
			}
		}
		wg.Done()
	}()
	wg.Wait()
	return
}

func deserialize(ipdatas, opdatas [][]byte) (inputs []In, outputs []Out) {
	inputs, outputs = make([]In, len(ipdatas)), make([]Out, len(opdatas))
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		for k, ipdata := range ipdatas {
			inputs[k] = InDecodeRLP(ipdata)
		}
		wg.Done()
	}()
	go func() {
		for k, opdata := range opdatas {
			outputs[k] = OutDecodeRLP(opdata)
		}
		wg.Done()
	}()
	wg.Wait()
	return inputs, outputs
}

// Transactions is a Transaction slice type for basic sorting.
type Transactions []*Transaction
