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
	"fmt"
	"math/big"

	"github.com/zipper-project/z0/common"
)

//ZipAccount .
type ZipAccount struct {
	balance *big.Int
	address common.Hash
}

func newZipAccount(balance *big.Int, address common.Hash) *ZipAccount {
	return &ZipAccount{balance: balance, address: address}
}

//MakeAsset .
func (z *ZipAccount) MakeAsset(value *big.Int, db int) Asset {
	// assets := make([]ZipAsset, 0)
	if z.balance.Cmp(value) < 0 {
		return nil
	}
	a := &ZipAsset{value, Zip, z.address, db}
	return a
}

//GetBalance .
func (z *ZipAccount) GetBalance() *big.Int {
	return z.balance
}

//ZipAsset .
type ZipAsset struct {
	value       *big.Int
	accountType uint
	address     common.Hash
	db          int
}

//Pay .
func (z *ZipAsset) Pay() error {
	zAccount := GetZipAccount(z.address, z.db).(*ZipAccount)
	if zAccount.balance.Cmp(z.value) < 0 {
		return fmt.Errorf("balance not enough")
	}
	zAccount.balance = new(big.Int).Sub(zAccount.balance, z.value)
	return nil
}

//Revenue .
func (z *ZipAsset) Revenue() {
	zAccount := GetZipAccount(z.address, z.db).(*ZipAccount)
	zAccount.balance = new(big.Int).Add(zAccount.balance, z.value)
}

//Encode .
func (z *ZipAsset) Encode() {

}
