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

package state

import (
	"bytes"
	"fmt"
	"io"

	"github.com/zipper-project/z0/common"
	"github.com/zipper-project/z0/crypto"
	"github.com/zipper-project/z0/utils/rlp"
)

var (
	emptyCodeHash = crypto.Keccak256(nil)
	emptyHash     = common.Hash{}
	emptyAdress   = common.Address{}
)

const (
	STROOTFlAG = 1
	ATROOTFlAG = 2
)

type Code []byte

func (self Code) String() string {
	return string(self) //strings.Join(Disassemble(self), " ")
}

func StorageCopy(origin map[common.Hash]common.Hash) map[common.Hash]common.Hash {
	new := make(map[common.Hash]common.Hash)
	for k, v := range origin {
		new[k] = v
	}
	return new
}

func CacheAccountCopy(origin map[string][]byte) map[string][]byte {
	new := make(map[string][]byte)
	for k, v := range origin {
		new[k] = v
	}
	return new
}

type stateObject struct {
	address  common.Address
	addrHash common.Hash // hash of  address of the account
	data     Account
	db       *StateDB

	dbErr error

	sttrie Trie // storage trie, which becomes non-nil on first access
	attrie Trie
	code   []byte

	cacheAccount map[string][]byte
	dirtyAccount map[string][]byte

	cachedStorage map[common.Hash]common.Hash // Storage entry cache to avoid duplicate reads
	dirtyStorage  map[common.Hash]common.Hash // Storage entries that need to be flushed to disk

	deleted   bool
	dirtyCode bool // true if the code was updated
}

func (s *stateObject) empty() bool {
	return bytes.Equal(s.data.AtRoot[:], emptyHash[:]) && bytes.Equal(s.data.CodeHash, emptyCodeHash) && len(s.cacheAccount) == 0
}

type Account struct {
	StRoot   common.Hash // merkle root of the storage trie
	AtRoot   common.Hash
	CodeHash []byte
}

func newObject(db *StateDB, address common.Address, data Account) *stateObject {
	if data.CodeHash == nil {
		data.CodeHash = emptyCodeHash
	}
	return &stateObject{
		db:            db,
		address:       address,
		addrHash:      crypto.Keccak256Hash(address[:]),
		data:          data,
		cacheAccount:  make(map[string][]byte),
		dirtyAccount:  make(map[string][]byte),
		cachedStorage: make(map[common.Hash]common.Hash),
		dirtyStorage:  make(map[common.Hash]common.Hash),
	}
}

func (c *stateObject) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, c.data)
}

func (self *stateObject) setError(err error) {
	if self.dbErr == nil {
		self.dbErr = err
	}
}

func (c *stateObject) getTrie(db Database, root common.Hash, flag int) Trie {
	var tr Trie
	var err error
	if flag == STROOTFlAG {
		if c.sttrie != nil {
			return c.sttrie
		}
		tr, err = db.OpenStorageTrie(c.addrHash, root)
		if err != nil {
			tr, _ = db.OpenStorageTrie(c.addrHash, common.Hash{})
			c.setError(fmt.Errorf("can't create storage trie: %v", err))
		}
		c.sttrie = tr
	} else {
		if c.attrie != nil {
			return c.attrie
		}
		tr, err = db.OpenStorageTrie(c.addrHash, root)
		if err != nil {
			tr, _ = db.OpenStorageTrie(c.addrHash, common.Hash{})
			c.setError(fmt.Errorf("can't create storage trie: %v", err))
		}
		c.attrie = tr
	}
	return tr
}

func (self *stateObject) GetState(db Database, key common.Hash) common.Hash {
	value, exists := self.cachedStorage[key]
	if exists {
		return value
	}
	// Load from DB in case it is missing.
	enc, err := self.getTrie(db, self.data.StRoot, STROOTFlAG).TryGet(key[:])
	if err != nil {
		self.setError(err)
		return common.Hash{}
	}

	if len(enc) > 0 {
		_, content, _, err := rlp.Split(enc)
		if err != nil {
			self.setError(err)
		}
		value.SetBytes(content)
		self.cachedStorage[key] = value
		return value
	}
	return common.Hash{}
}

// SetState updates a value in account storage.
func (self *stateObject) SetState(db Database, key, value common.Hash) {
	self.db.journal.append(storageChange{
		account:  &self.address,
		key:      key,
		prevalue: self.GetState(db, key),
	})
	self.setState(key, value)
}

func (self *stateObject) setState(key, value common.Hash) {
	self.cachedStorage[key] = value
	self.dirtyStorage[key] = value
}

func (self *stateObject) GetAccount(db Database, key string) []byte {
	value, exists := self.cacheAccount[key]
	if exists {
		return value
	}
	enc, err := self.getTrie(db, self.data.AtRoot, ATROOTFlAG).TryGet([]byte(key))
	if err != nil {
		self.setError(err)
		return []byte{}
	}
	self.cacheAccount[key] = enc
	return enc
}

func (self *stateObject) SetAccount(db Database, key string, value []byte) {
	self.db.journal.append(accountChange{
		account:  &self.address,
		key:      key,
		prevalue: self.GetAccount(db, key),
	})
	self.setAccount(key, value)
}

func (self *stateObject) setAccount(key string, value []byte) {
	self.cacheAccount[key] = value
	self.dirtyAccount[key] = value
}

func (self *stateObject) DeleteAccount(db Database, key string) {
	self.db.journal.append(accountChange{
		account:  &self.address,
		key:      key,
		prevalue: self.GetAccount(db, key),
	})
	self.deleteAccount(key)
}

func (self *stateObject) deleteAccount(key string) {
	delete(self.cacheAccount, key)
	delete(self.dirtyAccount, key)
}

func (self *stateObject) Code(db Database) []byte {
	if self.code != nil {
		return self.code
	}
	if bytes.Equal(self.CodeHash(), emptyCodeHash) {
		return nil
	}
	code, err := db.TrieDB().Node(common.BytesToHash(self.CodeHash()))
	if err != nil {
		self.setError(fmt.Errorf("can't load code hash %x: %v", self.CodeHash(), err))
	}
	self.code = code
	return code
}

func (self *stateObject) SetCode(codeHash common.Hash, code []byte) {
	prevcode := self.Code(self.db.db)
	self.db.journal.append(codeChange{
		account:  &self.address,
		prevhash: self.CodeHash(),
		prevcode: prevcode,
	})
	self.setCode(codeHash, code)
}

func (self *stateObject) setCode(codeHash common.Hash, code []byte) {
	self.code = code
	self.data.CodeHash = codeHash[:]
	self.dirtyCode = true
}

func (self *stateObject) CodeHash() []byte {
	return self.data.CodeHash
}

func (self *stateObject) updateStorageTrie(db Database) Trie {
	tr := self.getTrie(db, self.data.StRoot, STROOTFlAG)
	for k, _ := range self.dirtyStorage {
		value, _ := self.cachedStorage[k]
		if ok := bytes.Equal(value[:], []byte{}); ok {
			self.setError(tr.TryDelete(k[:]))
			continue
		}
		v, _ := rlp.EncodeToBytes(bytes.TrimLeft(value[:], "\x00"))
		self.setError(tr.TryUpdate(k[:], v))
	}
	return tr
}

func (self *stateObject) updateAccountTrie(db Database) Trie {
	tr := self.getTrie(db, self.data.AtRoot, ATROOTFlAG)
	for k, _ := range self.dirtyAccount {
		value, _ := self.cacheAccount[k]
		if ok := bytes.Equal(value, []byte{}); ok {
			self.setError(tr.TryDelete([]byte(k)))
			continue
		}
		self.setError(tr.TryUpdate([]byte(k), value))
	}
	return tr
}

func (self *stateObject) updateRoot(db Database) {
	self.updateStorageTrie(db)
	self.data.StRoot = self.sttrie.Hash()
	self.updateAccountTrie(db)
	self.data.AtRoot = self.attrie.Hash()
}

func (self *stateObject) CommitTrie(db Database) error {
	self.updateStorageTrie(db)
	if self.dbErr != nil {
		return self.dbErr
	}
	root, err := self.sttrie.Commit(nil)
	if err == nil {
		self.data.StRoot = root
	}
	self.updateAccountTrie(db)
	if self.dbErr != nil {
		return self.dbErr
	}
	root, err = self.attrie.Commit(nil)
	if err == nil {
		self.data.AtRoot = root
	}
	return err
}

func (self *stateObject) deepCopy(db *StateDB) *stateObject {
	stateObject := newObject(db, self.address, self.data)
	if self.sttrie != nil {
		stateObject.sttrie = db.db.CopyTrie(self.sttrie)
	}
	if self.sttrie != nil {
		stateObject.sttrie = db.db.CopyTrie(self.sttrie)
	}
	stateObject.dirtyStorage = StorageCopy(self.dirtyStorage)
	stateObject.cachedStorage = StorageCopy(self.cachedStorage)
	stateObject.dirtyAccount = CacheAccountCopy(self.dirtyAccount)
	stateObject.cacheAccount = CacheAccountCopy(self.cacheAccount)
	stateObject.deleted = self.deleted
	return stateObject
}

func (c *stateObject) Address() common.Address {
	return c.address
}
