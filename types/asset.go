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
	"math/big"
	"reflect"

	"github.com/zipper-project/z0/common"
)

var (
	//ZipAccount chain asset
	ZipAssetID = common.Address{1}
	//ZipAccount chain asset
	ZipAccount = common.Address{2}
)

var (
	AMInputType  = reflect.TypeOf(AMInput{})
	AMOutputType = reflect.TypeOf(AMOutput{})
	DefaultType  = reflect.TypeOf([]interface{}{})
)

const (
	// AccountModelType asset based account model
	AccountModelType uint8 = iota
)

type AccountModel struct {
}

type AMInput struct {
	AssertID *common.Address `json:"assertid"`
	Payload  []byte          `json:"payload"`
}

type AMOutput struct {
	AssertID *common.Address `json:"assertid"`
	Address  *common.Address `json:"to"`
	Value    *big.Int        `josn:"value"`
}
