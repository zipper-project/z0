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
	"encoding/json"
	"math/big"

	"github.com/zipper-project/z0/common"
	"github.com/zipper-project/z0/utils/rlp"
)

// In represents an asset input in the asset
type In interface {
	MarshalJSON() []byte
	EncodeRLP() []byte
}

// Out represents an asset output in the asset
type Out interface {
	MarshalJSON() []byte
	EncodeRLP() []byte
}

const (
	// AccountModelType asset based account model
	AccountModelType uint8 = iota
)

const (
	// InType input struct type
	InType uint8 = iota
	// OutType output struct type
	OutType
)

// InDecodeRLP decode assert
func InDecodeRLP(data []byte) In {
	var (
		result In
		err    error
	)
	switch uint8(data[0]) {
	case AccountModelType:
		in := new(AMInput)
		err = rlp.Decode(bytes.NewReader(data[1:]), in)
		result = in
	}
	if err != nil {
		panic(err)
	}
	return result
}

// OutDecodeRLP decode assert
func OutDecodeRLP(data []byte) Out {
	var (
		result Out
		err    error
	)
	switch uint8(data[0]) {
	case AccountModelType:
		out := new(AMOutput)
		err = rlp.Decode(bytes.NewReader(data[1:]), out)
		result = out
	}
	if err != nil {
		panic(err)
	}
	return result
}

type AccountModel struct {
}

type AMInput struct {
	Amount *big.Int `json:"amount"    gencodec:"required"`
}

func (a *AMInput) MarshalJSON() []byte {
	bytes, err := json.Marshal(a)
	if err != nil {
		panic(err)
	}
	return bytes
}
func (a *AMInput) EncodeRLP() []byte { return encodeRLP(a, AccountModelType) }

type AMOutput struct {
	Address *common.Address `json:"address"    gencodec:"required"`
}

func (a *AMOutput) MarshalJSON() []byte {
	bytes, err := json.Marshal(a)
	if err != nil {
		panic(err)
	}
	return bytes
}

func (a *AMOutput) EncodeRLP() []byte { return encodeRLP(a, AccountModelType) }

func encodeRLP(val interface{}, modelType uint8) []byte {
	bytes, err := rlp.EncodeToBytes(val)
	if err != nil {
		panic(err)
	}
	result := make([]byte, len(bytes)+1)
	result[0] = byte(modelType)
	copy(result[1:], bytes)
	return result
}
