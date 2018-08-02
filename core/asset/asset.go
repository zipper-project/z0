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
	// General account
	General = iota
	// Utxo account
	Utxo
)

var (
	assetlist = []byte("alist")
	assetType = []byte("aType")
)

//Asset operating user assets
type Asset struct {
	db StateDB
}

//NewAsset create Asset
func NewAsset(db StateDB) *Asset {
	return &Asset{db}
}

// RegisterAsset create asset
func (a *Asset) RegisterAsset(baseType int, accountAddr common.Address, nonce uint64, desc string) (common.Address, error) {
	var addr common.Address
	var err error
	switch baseType {
	case General:
		addr, err = registerGeneralAsset(a.db, accountAddr, nonce, desc)
		if err != nil {
			return addr, err
		}
	case Utxo:
		fmt.Println("Utxo")
	}
	return addr, nil
}

// IssueAsset issue asset
func (a *Asset) IssueAsset(baseType int, targetAddr common.Address, assetAddr common.Address, value interface{}) error {
	switch baseType {
	case General:
		v := value.(*big.Int)
		err := issueGeneralAsset(a.db, targetAddr, assetAddr, v)
		if err != nil {
			return err
		}
	case Utxo:
		fmt.Println("Utxo")
	}
	return nil
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

//GetAccountList .
func (a *Asset) GetAccountList(baseType int, address common.Address) ([]common.Address, error) {
	key := address.String() + string(assetlist)
	v := a.db.GetAccount(address, key)

	var list []common.Address
	if !bytes.Equal(v, []byte{}) {
		err := rlp.Decode(bytes.NewReader(v), &list)
		if err != nil {
			return nil, fmt.Errorf("Error: %v", err)
		}
		return list, nil
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
func (a *Asset) SubBalance(baseType int, targetAddr common.Address, assetAddr common.Address, value interface{}) error {
	switch baseType {
	case General:
		v := value.(*big.Int)
		err := subGeneralBalance(a.db, targetAddr, assetAddr, v)
		if err != nil {
			return err
		}
	case Utxo:
		fmt.Println("Utxo")
	}
	return nil
}

// AddBalance add account balance
func (a *Asset) AddBalance(baseType int, targetAddr common.Address, assetAddr common.Address, value interface{}) error {
	switch baseType {
	case General:
		v := value.(*big.Int)
		err := addGeneralBalance(a.db, targetAddr, assetAddr, v)
		if err != nil {
			return err
		}
	case Utxo:
		fmt.Println("Utxo")
	}
	return nil
}

// GetBalance get account balance
func (a *Asset) GetBalance(baseType int, targetAddr common.Address, assetAddr common.Address) (interface{}, error) {
	switch baseType {
	case General:
		balance, err := getGeneralBalance(a.db, targetAddr, assetAddr)
		if err != nil {
			return nil, err
		}
		return balance, nil
	case Utxo:
		fmt.Println("Utxo")
	}
	return nil, nil
}

// GetNonce get nonce
func (a *Asset) GetNonce(targetAddr common.Address) (uint64, error) {
	accountByte := a.db.GetAccount(targetAddr, targetAddr.String())
	var account Account
	if !bytes.Equal(accountByte, []byte{}) {
		err := rlp.Decode(bytes.NewReader(accountByte), &account)
		if err != nil {
			return 0, err
		}
	} else {
		return 0, fmt.Errorf("account not exit")
	}
	return account.Nonce, nil
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

// //Asset base asset interface
// type Asset interface {
// 	Pay() error
// 	Revenue() error
// }

// //Account base account interface
// type Account interface {
// 	MakeAsset(value *big.Int, db state.StateDB) Asset
// 	GetBalance() *big.Int
// }

// //GetAccountType get all account type of the user
// func GetAccountType(address common.Address, statedb state.StateDB) ([]uint, error) {
// 	key := address.String() + string(accountType)
// 	val := statedb.GetAccount(address, key)
// 	var types []uint
// 	err := rlp.Decode(bytes.NewReader(val), &types)
// 	if err != nil {
// 		return nil, fmt.Errorf("Error: %v", err)
// 	}
// 	return types, nil
// }

// //GetZipAccount get zipaccount information of the user
// func GetZipAccount(address common.Address, statedb state.StateDB) (*ZipAccount, error) {
// 	key := address.String() + string(account) + strconv.Itoa(Zip)

// 	val := statedb.GetAccount(address, key)
// 	var account ZipAccount
// 	err := rlp.Decode(bytes.NewReader(val), &account)
// 	if err != nil {
// 		return nil, fmt.Errorf("Error: %v", err)
// 	}

// 	return newZipAccount(account.Balance, address), nil
// }
