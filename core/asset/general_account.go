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
	"strconv"
	"strings"

	"github.com/zipper-project/z0/common"
	"github.com/zipper-project/z0/crypto"
	"github.com/zipper-project/z0/utils/rlp"
)

type genAssetInfo struct {
	Name         string
	Total        *big.Int
	DecimalPoint uint64
	Balance      *big.Int
}

func registerGeneralAsset(db StateDB, accountAddr common.Address, nonce uint64, desc string) (common.Address, error) {
	infoArray := strings.Split(desc, ",")
	if len(infoArray) != 3 {
		return common.Address{}, fmt.Errorf("RegisterGeneralAsset DATA INVALID")
	}
	name := infoArray[0]
	total, err := strconv.ParseInt(infoArray[1], 10, 64)
	if err != nil {
		return common.Address{}, fmt.Errorf("RegisterGeneralAsset total INVALID")
	}
	decimalPoint, err := strconv.ParseUint(infoArray[2], 10, 32)
	if err != nil {
		return common.Address{}, fmt.Errorf("RegisterGeneralAsset decimalPoint INVALID")
	}
	info := &genAssetInfo{name, big.NewInt(total), decimalPoint, big.NewInt(total)}
	b := new(bytes.Buffer)
	err = rlp.Encode(b, info)
	if err != nil {
		return common.Address{}, err
	}

	assetAddress := crypto.CreateAssetAddress(accountAddr, nonce, name)
	db.SetAccount(assetAddress, assetAddress.String(), b.Bytes())

	//save
	assetTypeKey := assetAddress.String() + string(assetType)
	b = new(bytes.Buffer)
	err = rlp.Encode(b, uint64(General))
	if err != nil {
		return common.Address{}, err
	}
	db.SetAccount(assetAddress, assetTypeKey, b.Bytes())

	return assetAddress, nil
}

func issueGeneralAsset(db StateDB, targetAddr common.Address, assetAddr common.Address, value *big.Int) error {
	//get target user  balance
	key := targetAddr.String() + assetAddr.String()
	selfAsset := db.GetAccount(targetAddr, key)
	var balance *big.Int
	if !bytes.Equal(selfAsset, []byte{}) {
		err := rlp.Decode(bytes.NewReader(selfAsset), &balance)
		if err != nil {
			return err
		}
	} else {
		balance = big.NewInt(0)
		setAccountList(General, db, targetAddr, assetAddr)
	}

	//get asset balance
	var info genAssetInfo
	asset := db.GetAccount(assetAddr, assetAddr.String())
	if !bytes.Equal(asset, []byte{}) {
		err := rlp.Decode(bytes.NewReader(asset), &info)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Asset not exit")
	}

	if info.Balance.Cmp(value) < 0 {
		return fmt.Errorf("Asset not enough")
	}

	info.Balance = new(big.Int).Sub(info.Balance, value)
	balance = new(big.Int).Add(balance, value)

	b := new(bytes.Buffer)
	err := rlp.Encode(b, balance)
	if err != nil {
		return err
	}
	//save target user  balance
	db.SetAccount(targetAddr, key, b.Bytes())

	b = new(bytes.Buffer)
	err = rlp.Encode(b, &info)
	if err != nil {
		return err
	}
	//save asset balance
	db.SetAccount(assetAddr, assetAddr.String(), b.Bytes())

	return nil
}

func subGeneralBalance(db StateDB, targetAddr common.Address, assetAddr common.Address, value *big.Int) error {
	key := targetAddr.String() + assetAddr.String()
	selfAsset := db.GetAccount(targetAddr, key)
	var balance *big.Int

	if !bytes.Equal(selfAsset, []byte{}) {
		err := rlp.Decode(bytes.NewReader(selfAsset), &balance)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Asset not exit")
	}

	if balance.Cmp(value) < 0 {
		return fmt.Errorf("Asset not enough")
	}
	balance = new(big.Int).Sub(balance, value)
	b := new(bytes.Buffer)
	err := rlp.Encode(b, &balance)
	if err != nil {
		return err
	}
	db.SetAccount(targetAddr, key, b.Bytes())

	return nil
}

func addGeneralBalance(db StateDB, targetAddr common.Address, assetAddr common.Address, value *big.Int) error {
	key := targetAddr.String() + assetAddr.String()
	selfAsset := db.GetAccount(targetAddr, key)
	var balance *big.Int

	if !bytes.Equal(selfAsset, []byte{}) {
		err := rlp.Decode(bytes.NewReader(selfAsset), &balance)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Asset not exit")
	}

	balance = new(big.Int).Add(balance, value)

	b := new(bytes.Buffer)
	err := rlp.Encode(b, balance)
	if err != nil {
		return err
	}
	db.SetAccount(targetAddr, key, b.Bytes())

	return nil
}

func getGeneralBalance(db StateDB, targetAddr common.Address, assetAddr common.Address) (*big.Int, error) {
	key := targetAddr.String() + assetAddr.String()
	selfAsset := db.GetAccount(targetAddr, key)
	var balance *big.Int

	if !bytes.Equal(selfAsset, []byte{}) {
		err := rlp.Decode(bytes.NewReader(selfAsset), &balance)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("Asset not exit")
	}

	return balance, nil
}

// //ZipAccount .
// type ZipAccount struct {
// 	Balance *big.Int
// 	address common.Address
// }

// func newZipAccount(balance *big.Int, address common.Address) *ZipAccount {
// 	return &ZipAccount{Balance: balance, address: address}
// }

// //MakeAsset .
// func (z *ZipAccount) MakeAsset(assetValue *big.Int, db state.StateDB) Asset {
// 	// assets := make([]ZipAsset, 0)
// 	if z.Balance.Cmp(assetValue) < 0 {
// 		return nil
// 	}
// 	zAsset := &ZipAsset{assetValue, Zip, z.address, db}
// 	return zAsset
// }

// //GetBalance .
// func (z *ZipAccount) GetBalance() *big.Int {
// 	return z.Balance
// }

// //ZipAsset .
// type ZipAsset struct {
// 	value       *big.Int
// 	accountType uint
// 	address     common.Address
// 	db          state.StateDB
// }

// //Pay .
// func (z *ZipAsset) Pay() error {
// 	zAccount, err := GetZipAccount(z.address, z.db)
// 	if err != nil {
// 		return err
// 	}
// 	// zAccount := account.(*ZipAccount)
// 	if zAccount.Balance.Cmp(z.value) < 0 {
// 		return fmt.Errorf("balance not enough")
// 	}
// 	zAccount.Balance = new(big.Int).Sub(zAccount.Balance, z.value)
// 	b := new(bytes.Buffer)
// 	err = rlp.Encode(b, zAccount)
// 	if err != nil {
// 		return err
// 	}
// 	key := z.address.String() + string(z.accountType)
// 	z.db.SetAccount(z.address, key, b.Bytes())
// 	return nil
// }

// //Revenue .
// func (z *ZipAsset) Revenue() error {
// 	zAccount, err := GetZipAccount(z.address, z.db)
// 	if err != nil {
// 		return err
// 	}
// 	// zAccount := account.(*ZipAccount)
// 	zAccount.Balance = new(big.Int).Add(zAccount.Balance, z.value)
// 	b := new(bytes.Buffer)
// 	err = rlp.Encode(b, zAccount)
// 	if err != nil {
// 		return err
// 	}
// 	key := z.address.String() + string(z.accountType)
// 	z.db.SetAccount(z.address, key, b.Bytes())
// 	return nil
// }
