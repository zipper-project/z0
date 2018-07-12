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

package asset

import (
	"math/big"
	"strconv"

	"github.com/z0/common"
)

const (
	// Zip account
	Zip = iota
	// Utxo account
	Utxo
)

var (
	account = []byte("account")
)

//Asset base asset interface
type Asset interface {
	Pay() error
	Revenue()
	Encode()
}

//Account base account interface
type Account interface {
	MakeAsset(value *big.Int, db int) Asset
	GetBalance() *big.Int
}

//GetAccountType get all account type of the user
func GetAccountType(address common.Hash, statedb interface{}) []int {
	// address = nil
	types := make([]int, 0)
	types = append(types, Zip)
	types = append(types, Utxo)
	return types
}

//GetZipAccount get zipaccount information of the user
func GetZipAccount(address common.Hash, statedb interface{}) Account {
	// address = nil
	strType := strconv.Itoa(Zip)
	key := make([]byte, len(address[:])+len(account)+len(strType[:]))
	copy(key, address[:])
	copy(key, account)
	copy(key, strType[:])

	balance := big.NewInt(1)

	return newZipAccount(balance, address)
}
