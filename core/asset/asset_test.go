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
	"encoding/json"
	"math/big"
	"testing"

	"github.com/zipper-project/z0/common"
	"github.com/zipper-project/z0/state"
	"github.com/zipper-project/z0/utils/zdb"
)

func TestGeneralAsset(t *testing.T) {
	db := zdb.NewMemDatabase()
	tridb := state.NewDatabase(db)
	statedb, err := state.New(common.Hash{}, tridb)

	account1 := common.Address{10}
	account100 := common.Address{100}
	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}
	asset := NewAsset(statedb)
	err = asset.CreateAccount(account1)
	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}

	info := &AccountAssetInfo{
		Name:     "test",
		Symbol:   "BTC",
		Total:    big.NewInt(2100),
		Decimals: 8,
		Owner:    account100}

	b, err := json.Marshal(info)
	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}

	aAddress, err := asset.RegisterAsset(AccountModel, account1, string(b))
	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}
	err = asset.IssueAsset(account100, aAddress, big.NewInt(20))
	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}

	v := asset.GetBalance(account100, aAddress)
	value := v.(*big.Int)
	if value.Cmp(big.NewInt(2120)) != 0 {
		t.Errorf("balace error 2120")
	}

	_, err = asset.GetUserAssets(account100)
	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}
	// fmt.Printf("type:%v address:%v name:%v balance:%v\n", assets[0].baseType, assets[0].assetAddr, assets[0].assetName, assets[0].balance)
	// if strings.Compare(aAddress.String(), list[0].String()) != 0 {
	// 	t.Errorf("GetAccountList error ")
	// }
	// fmt.Printf("asset address:%v\n", aAddress)
	// fmt.Printf("list         :%v\n", list[0])

	err = asset.SubBalance(account100, aAddress, big.NewInt(1))
	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}
	v = asset.GetBalance(account100, aAddress)
	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}
	value = v.(*big.Int)
	// fmt.Printf("get value:%v\n", value)
	if value.Cmp(big.NewInt(2119)) != 0 {
		t.Errorf("balace error 2120")
	}

	err = asset.AddBalance(account100, aAddress, big.NewInt(3))
	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}
	v = asset.GetBalance(account100, aAddress)
	value = v.(*big.Int)
	// fmt.Printf("get value:%v\n", value)
	if value.Cmp(big.NewInt(2122)) != 0 {
		t.Errorf("balace error 2122")
	}
	// err = asset.AddBalance(common.Address{2}, aAddress, big.NewInt(60))
	// if err != nil {
	// 	t.Errorf("Unexpected error : %v", err)
	// }
	// v, err = asset.GetBalance(common.Address{2}, aAddress)
	// value = v.(*big.Int)
	// fmt.Printf("get value:%v\n", value)

	boo, err := asset.EnoughBalance(account100, aAddress, big.NewInt(2109))
	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}
	if !boo {
		t.Errorf("EnoughBalance error ")
	}
	// fmt.Printf("EnoughBalance:%v\n", boo)

	err = asset.CreateAccount(account1)
	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}
	nonce := asset.GetNonce(account1)
	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}
	if nonce != 0 {
		t.Errorf("GetNonce error ")
	}
	// fmt.Printf("get nonce:%v\n", v1)
	err = asset.SetNonce(account1, 1)
	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}

	nonce = asset.GetNonce(account1)
	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}
	if nonce != 1 {
		t.Errorf("SetNonce error ")
	}
	// fmt.Printf("get nonce:%v\n", v1)

	boo = asset.Empty(common.Address{2})
	if !boo {
		t.Errorf("Empty error ")
	}
	// fmt.Printf("Empty:%v\n", boo)

	boo = asset.Exist(common.Address{2})
	if boo {
		t.Errorf("Empty error ")
	}
	// fmt.Printf("Exist:%v\n", boo)

	InitZip(statedb, big.NewInt(1000), 8)
	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}

	err = asset.SubBalance(ZIPACCOUNT, ZIPASSET, big.NewInt(1))
	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}
	v = asset.GetBalance(ZIPACCOUNT, ZIPASSET)
	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}
	value = v.(*big.Int)
	if value.Cmp(big.NewInt(999)) != 0 {
		t.Errorf("balace error 999")
	}
	_, err = asset.GetUserAssets(ZIPACCOUNT)
	if err != nil {
		t.Errorf("Unexpected error : %v", err)
	}
	// fmt.Printf("type:%v address:%v name:%v balance:%v\n", assets[0].baseType, assets[0].assetAddr, assets[0].assetName, assets[0].balance)
}
