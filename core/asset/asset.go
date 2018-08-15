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
	"bytes"
	"fmt"
	"math/big"

	"github.com/zipper-project/z0/common"
	"github.com/zipper-project/z0/utils/rlp"
)

const (
	// AccountModel account
	AccountModel = iota
	// UtxoModel account
	UtxoModel
)

var (
	assetlist = []byte("alist")
	assetType = []byte("aType")
)

var (
	//ZIPASSET chain asset
	ZIPASSET = common.Address{1}
	//ZIPACCOUNT chain asset
	ZIPACCOUNT = common.Address{2}
)

//Asset operating user assets
type Asset struct {
	db StateDB
}

//NewAsset create Asset
func NewAsset(db StateDB) *Asset {
	return &Asset{db}
}

//InitZip Genesis asset ZIP
func InitZip(db StateDB, total *big.Int, decimals uint64) error {
	asset := db.GetAccount(ZIPASSET, ZIPASSET.String())
	if !bytes.Equal(asset, []byte{}) {
		return nil
	}
	info := &AccountAssetInfo{
		Name:     "zipper",
		Symbol:   "ZIP",
		Total:    total,
		Decimals: decimals,
		Owner:    ZIPACCOUNT}
	//save zip asset info
	b := new(bytes.Buffer)
	err := rlp.Encode(b, &info)
	if err != nil {
		return err
	}
	assetAddr := ZIPASSET
	db.SetAccount(assetAddr, assetAddr.String(), b.Bytes())

	//save base type
	assetTypeKey := assetAddr.String() + string(assetType)
	b = new(bytes.Buffer)
	err = rlp.Encode(b, uint64(AccountModel))
	if err != nil {
		return err
	}
	db.SetAccount(assetAddr, assetTypeKey, b.Bytes())
	//issue balance to owner
	err = setAccountList(AccountModel, db, info.Owner, assetAddr)
	if err != nil {
		return err
	}
	key := info.Owner.String() + assetAddr.String()
	b = new(bytes.Buffer)
	err = rlp.Encode(b, info.Total)
	if err != nil {
		return err
	}
	db.SetAccount(info.Owner, key, b.Bytes())
	return nil
}

// RegisterAsset create asset
func (a *Asset) RegisterAsset(baseType int, accountAddr common.Address, desc string) (common.Address, error) {
	var addr common.Address
	var err error
	switch baseType {
	case AccountModel:
		addr, err = registerAccountAsset(a.db, accountAddr, a.GetNonce(accountAddr), desc)
		if err != nil {
			return addr, err
		}
	case UtxoModel:
		fmt.Println("Utxo")
	}
	return addr, nil
}

// SetNewOwner .
func (a *Asset) SetNewOwner(oldOwner common.Address, assetAddr common.Address, newOwner common.Address) (bool, error) {
	baseType, err := a.getAssetType(assetAddr)
	if err != nil {
		return false, err
	}
	var ok bool
	switch baseType {
	case AccountModel:
		ok, err = setAccountNewOwner(a.db, oldOwner, assetAddr, newOwner)
		if err != nil {
			return false, err
		}
	case UtxoModel:
		fmt.Println("Utxo")
	}
	return ok, nil
}

// IssueAsset issue asset
func (a *Asset) IssueAsset(ownerAddr common.Address, assetAddr common.Address, value interface{}) error {
	baseType, err := a.getAssetType(assetAddr)
	if err != nil {
		return err
	}
	switch baseType {
	case AccountModel:
		v := value.(*big.Int)
		err := issueAccountAsset(a.db, ownerAddr, assetAddr, v)
		if err != nil {
			return err
		}
	case UtxoModel:
		fmt.Println("Utxo")
	}
	return nil
}

//UserAsset user asset info
type UserAsset struct {
	baseType  uint
	assetAddr common.Address
	assetName string
	balance   *big.Int
}

//GetUserAssets .
func (a *Asset) GetUserAssets(address common.Address) ([]UserAsset, error) {
	key := address.String() + string(assetlist)
	v := a.db.GetAccount(address, key)

	var list []common.Address
	if !bytes.Equal(v, []byte{}) {
		err := rlp.Decode(bytes.NewReader(v), &list)
		if err != nil {
			return nil, fmt.Errorf("Error: %v", err)
		}

		assets := make([]UserAsset, 0)
		for _, assetAddr := range list {
			baseType, err := a.getAssetType(assetAddr)
			if err != nil {
				return nil, err
			}
			switch baseType {
			case AccountModel:
				balance, err := getAccountBalance(a.db, address, assetAddr)
				if err != nil {
					return nil, err
				}
				info, err := getAccountAssetInfo(a.db, assetAddr)
				if err != nil {
					return nil, err
				}
				asset := UserAsset{
					baseType:  AccountModel,
					assetAddr: assetAddr,
					assetName: info.Symbol,
					balance:   balance}
				assets = append(assets, asset)
			case UtxoModel:
				fmt.Println("Utxo")
			}
		}
		return assets, nil
	}

	return nil, fmt.Errorf("not Account list info")
}

//Account .
type Account struct {
	Nonce uint64
}

// CreateAccount create account
func (a *Asset) CreateAccount(addr common.Address) error {
	account := &Account{0}
	b := new(bytes.Buffer)
	err := rlp.Encode(b, account)
	if err != nil {
		return err
	}
	a.db.SetAccount(addr, addr.String(), b.Bytes())
	return nil
}

// SubBalance sub account balance
func (a *Asset) SubBalance(targetAddr common.Address, assetAddr common.Address, value interface{}) error {
	baseType, err := a.getAssetType(assetAddr)
	if err != nil {
		return err
	}
	switch baseType {
	case AccountModel:
		v := value.(*big.Int)
		err := subAccountBalance(a.db, targetAddr, assetAddr, v)
		if err != nil {
			return err
		}
	case UtxoModel:
		fmt.Println("Utxo")
	}
	return nil
}

// AddBalance add account balance
func (a *Asset) AddBalance(targetAddr common.Address, assetAddr common.Address, value interface{}) error {
	baseType, err := a.getAssetType(assetAddr)
	if err != nil {
		return err
	}
	switch baseType {
	case AccountModel:
		v := value.(*big.Int)
		err := addAccountlBalance(a.db, targetAddr, assetAddr, v)
		if err != nil {
			return err
		}
	case UtxoModel:
		fmt.Println("Utxo")
	}
	return nil
}

// EnoughBalance .
func (a *Asset) EnoughBalance(targetAddr common.Address, assetAddr common.Address, value interface{}) (bool, error) {
	baseType, err := a.getAssetType(assetAddr)
	if err != nil {
		return false, err
	}
	var enough bool
	switch baseType {
	case AccountModel:
		v := value.(*big.Int)
		enough, err = enoughAccountBalance(a.db, targetAddr, assetAddr, v)
		if err != nil {
			return false, err
		}
	case UtxoModel:
		fmt.Println("Utxo")
	}
	return enough, nil
}

// GetBalance get account balance
func (a *Asset) GetBalance(targetAddr common.Address, assetAddr common.Address) interface{} {
	baseType, err := a.getAssetType(assetAddr)
	if err != nil {
		panic("GetBalance error")
	}
	switch baseType {
	case AccountModel:
		balance, err := getAccountBalance(a.db, targetAddr, assetAddr)
		if err != nil {
			panic("GetBalance error")
		}
		return balance
	case UtxoModel:
		fmt.Println("Utxo")
	}
	return nil
}

// GetNonce get nonce
func (a *Asset) GetNonce(targetAddr common.Address) uint64 {
	accountByte := a.db.GetAccount(targetAddr, targetAddr.String())
	var account Account
	if !bytes.Equal(accountByte, []byte{}) {
		err := rlp.Decode(bytes.NewReader(accountByte), &account)
		if err != nil {
			panic("GetNonce error")
		}
	} else {
		panic("GetNonce error")
	}
	return account.Nonce
}

// SetNonce set nonce
func (a *Asset) SetNonce(targetAddr common.Address, nonce uint64) error {
	accountByte := a.db.GetAccount(targetAddr, targetAddr.String())
	var account Account
	if !bytes.Equal(accountByte, []byte{}) {
		err := rlp.Decode(bytes.NewReader(accountByte), &account)
		if err != nil {
			return err
		}
	} else {
		return nil
	}

	account.Nonce = nonce
	b := new(bytes.Buffer)
	err := rlp.Encode(b, &account)
	if err != nil {
		return err
	}
	a.db.SetAccount(targetAddr, targetAddr.String(), b.Bytes())
	return nil
}

// Empty returns whether the account is empty
func (a *Asset) Empty(targetAddr common.Address) bool {
	accountByte := a.db.GetAccount(targetAddr, targetAddr.String())
	var account Account
	if !bytes.Equal(accountByte, []byte{}) {
		err := rlp.Decode(bytes.NewReader(accountByte), &account)
		if err != nil {
			return true
		}
		if account.Nonce == 0 {
			return true
		}
	} else {
		return true
	}

	return false
}

// Exist returns whether the account is exist
func (a *Asset) Exist(targetAddr common.Address) bool {
	accountByte := a.db.GetAccount(targetAddr, targetAddr.String())
	var account Account
	if !bytes.Equal(accountByte, []byte{}) {
		err := rlp.Decode(bytes.NewReader(accountByte), &account)
		if err != nil {
			return false
		}
	} else {
		return false
	}

	return true
}

//
func (a *Asset) getAssetType(assetAddr common.Address) (int, error) {
	assetTypeKey := assetAddr.String() + string(assetType)
	v := a.db.GetAccount(assetAddr, assetTypeKey)

	var utype uint64
	if !bytes.Equal(v, []byte{}) {
		err := rlp.Decode(bytes.NewReader(v), &utype)
		if err != nil {
			return -1, fmt.Errorf("Error: %v", err)
		}
	} else {
		return -1, fmt.Errorf("Asset not create")
	}
	return int(utype), nil
}

func setAccountList(baseType int, statedb StateDB, address common.Address, assetAddr common.Address) error {
	key := address.String() + string(assetlist)
	v := statedb.GetAccount(address, key)

	var list []common.Address
	if !bytes.Equal(v, []byte{}) {
		err := rlp.Decode(bytes.NewReader(v), &list)
		if err != nil {
			return fmt.Errorf("Error: %v", err)
		}

		for _, t := range list {
			if bytes.Equal(t.Bytes(), assetAddr.Bytes()) {
				return nil
			}
		}
	}
	list = append(list, assetAddr)
	b := new(bytes.Buffer)
	err := rlp.Encode(b, list)
	if err != nil {
		return err
	}
	statedb.SetAccount(address, key, b.Bytes())
	return nil
}
