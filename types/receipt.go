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
	"math/big"

	"github.com/zipper-project/z0/common"
	"github.com/zipper-project/z0/utils/rlp"
)

var (
	receiptStatusFailedRLP     = []byte{}
	receiptStatusSuccessfulRLP = []byte{0x01}
)

const (
	// ReceiptStatusFailed is the status code of a transaction if execution failed.
	ReceiptStatusFailed = uint64(0)

	// ReceiptStatusSuccessful is the status code of a transaction if execution succeeded.
	ReceiptStatusSuccessful = uint64(1)
)

// InternalTx represents the results of a contract internal transaction.
type InternalTx struct {
	OpCode byte           `json:"opcode"        `
	From   common.Address `json:"from"        `
	To     common.Address `json:"to"        `
	Value  *big.Int       `json:"value"        `
}

// Receipt represents the results of a transaction.
type Receipt struct {
	// Consensus fields
	PostState         []byte        `json:"root"`
	Status            uint64        `json:"status"`
	Internal          []*InternalTx `json:"internal" `
	CumulativeGasUsed uint64        `json:"cumulativeGasUsed"`
	Bloom             Bloom         `json:"logsBloom"        `
	Logs              []*Log        `json:"logs"             `

	// Implementation fields (don't reorder!)
	TxHash          common.Hash    `json:"transactionHash"`
	ContractAddress common.Address `json:"contractAddress"`
	GasUsed         uint64         `json:"gasUsed"`
}

// NewReceipt creates a barebone transaction receipt, copying the init fields.
func NewReceipt(root []byte, failed bool, cumulativeGasUsed uint64) *Receipt {
	r := &Receipt{PostState: common.CopyBytes(root), CumulativeGasUsed: cumulativeGasUsed}
	if failed {
		r.Status = ReceiptStatusFailed
	} else {
		r.Status = ReceiptStatusSuccessful
	}
	return r
}

// EncodeRLP implements rlp.Encoder
func (r *Receipt) EncodeRLP() ([]byte, error) {
	return rlp.EncodeToBytes(r)
}

// DecodeRLP implements rlp.Decoder
func (r *Receipt) DecodeRLP(data []byte) error {
	return rlp.DecodeBytes(data, r)
}

// Size returns the approximate memory used by all internal contents
func (r *Receipt) Size() common.StorageSize {
	bytes, _ := r.EncodeRLP()
	return common.StorageSize(len(bytes))
}

// Receipts is a wrapper around a Receipt array to implement DerivableList.
type Receipts []*Receipt

// Len returns the number of receipts in this list.
func (r Receipts) Len() int { return len(r) }

// GetRlp returns the RLP encoding of one receipt from the list.
func (r Receipts) GetRlp(i int) []byte {
	bytes, err := rlp.EncodeToBytes(r[i])
	if err != nil {
		panic(err)
	}
	return bytes
}
