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
	"testing"

	"github.com/zipper-project/z0/common"
	"github.com/zipper-project/z0/state"
	"github.com/zipper-project/z0/zdb"
)

func TestGeneralAsset(t *testing.T) {
	db := zdb.NewMemDatabase()
	tridb := state.NewDatabase(db)
	statedb, err := state.New(common.Hash{}, tridb)

	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}

	asset := NewAsset(statedb)
	aAddress, err := asset.RegisterAsset(General, common.Address{1}, 1, "test,1000000000,8")
	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}
	err = asset.IssueAsset(General, common.Address{1}, aAddress, big.NewInt(10))
	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}

	err = asset.IssueAsset(General, common.Address{1}, aAddress, big.NewInt(30))
	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}

	err = asset.IssueAsset(General, common.Address{1}, aAddress, big.NewInt(130))
	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}

	v, err := asset.GetBalance(General, common.Address{1}, aAddress)
	value := v.(*big.Int)
	fmt.Printf("value:%v\n", value)

	list, err := asset.GetAccountList(General, common.Address{1})
	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}
	fmt.Printf("asset address:%v\n", aAddress)
	fmt.Printf("list         :%v\n", list[0])

	err = asset.SubBalance(General, common.Address{1}, aAddress, big.NewInt(30))
	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}
	v, err = asset.GetBalance(General, common.Address{1}, aAddress)
	value = v.(*big.Int)
	fmt.Printf("get value:%v\n", value)

	err = asset.AddBalance(General, common.Address{1}, aAddress, big.NewInt(60))
	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}
	v, err = asset.GetBalance(General, common.Address{1}, aAddress)
	value = v.(*big.Int)
	fmt.Printf("get value:%v\n", value)

	err = asset.CreateAccount(common.Address{1})
	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}
	v1, err := asset.GetNonce(common.Address{1})
	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}
	fmt.Printf("get nonce:%v\n", v1)
	err = asset.SetNonce(common.Address{1}, 1)
	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}

	v1, err = asset.GetNonce(common.Address{1})
	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}
	fmt.Printf("get nonce:%v\n", v1)

	boo := asset.Empty(common.Address{2})
	fmt.Printf("Empty:%v\n", boo)

	boo = asset.Exist(common.Address{2})
	fmt.Printf("Exist:%v\n", boo)
}
