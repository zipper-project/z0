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
	"testing"
	"fmt"
	"strconv"

	"github.com/zipper-project/z0/common"
	"github.com/zipper-project/z0/zdb"
)

func TestUpdateLeaks(t *testing.T) {
	db := zdb.NewMemDatabase()
	tridb := NewDatabase(db)
	state, _ := New(common.Hash{}, tridb)

	var id int
	for i := byte(0); i < 4; i++ {
		addr := common.BytesToAddress([]byte{i})
		//addr[len(addr)-2] = i

		if int(i) == 1 {
			id = state.Snapshot()
		}
		for j := 0; j < 3; j++ {
			data := []byte("stk" + strconv.Itoa(int(i)) + strconv.Itoa(int(j)))
			value := []byte("stv" + strconv.Itoa(int(i)) + strconv.Itoa(int(j)))
			state.SetState(addr, common.BytesToHash(data), common.BytesToHash(value))

			key := "at" + strconv.Itoa(int(i)) + strconv.Itoa(int(j))
			value = []byte("account" + strconv.Itoa(int(i)) + strconv.Itoa(int(j)))
			state.SetAccount(addr, key, value)
		}
		d := []byte("code" + strconv.Itoa(int(i)))
		state.SetCode(addr, d)
		if int(i) == 1 {
			state.RevertToSnapshot(id)
		}
		state.IntermediateRoot(false)
	}

	Root, _ := state.Commit(true)
	tridb.TrieDB().Commit(Root, false)
	cpy, _ := New(Root, tridb)
	for i := byte(0); i < 4; i++ {
		addr := common.BytesToAddress([]byte{i})

		for j := 0; j < 3; j++ {
			data := []byte("stk" + strconv.Itoa(int(i)) + strconv.Itoa(int(j)))
			value := []byte("stv" + strconv.Itoa(int(i)) + strconv.Itoa(int(j)) + "new")

			cpy.SetState(addr, common.BytesToHash(data), common.BytesToHash(value))

			key := "at" + strconv.Itoa(int(i)) + strconv.Itoa(int(j))
			value = []byte("account" + strconv.Itoa(int(i)) + strconv.Itoa(int(j)) + "new")
			cpy.SetAccount(addr, key, value)
		}
		d := []byte("code" + strconv.Itoa(int(i)) + "new")
		state.SetCode(addr, d)
		cpy.IntermediateRoot(false)
	}
	cpyRoot, _ := cpy.Commit(true)
	tridb.TrieDB().Commit(cpyRoot, false)

	cpyt, _ := New(cpyRoot, tridb)
	for i := byte(0); i < 4; i++ {
		addr := common.BytesToAddress([]byte{i})

		for j := 0; j < 3; j++ {
			data := []byte("stk" + strconv.Itoa(int(i)) + strconv.Itoa(int(j)))
			t := cpyt.GetState(addr, common.BytesToHash(data))
			fmt.Println(string(t[:]))

		}
		for j := 0; j < 3; j++ {
			key := "at" + strconv.Itoa(int(i)) + strconv.Itoa(int(j))
			m := cpyt.GetAccount(addr, key)
			fmt.Println(string(m))
		}

		code := cpyt.GetCode(addr)
		fmt.Println(string(code))
	}
}

func newtridb() Database {
	return NewDatabase(zdb.NewMemDatabase())
}

func TestSetState(t *testing.T) {
	tridb := newtridb()
	state, _ := New(common.Hash{}, tridb)
	for i := 0; i < 4; i++ {
		addr := common.BytesToAddress([]byte{byte(i)})
		for j := 0; j < 4; j++ {
			key := []byte("sk" + strconv.Itoa(i) + strconv.Itoa(j))
			value := []byte("sv" + strconv.Itoa(i) + strconv.Itoa(j))
			state.SetState(addr, common.BytesToHash(key), common.BytesToHash(value))
		}
		state.IntermediateRoot(true)
	}
	root, err := state.Commit(true)
	if err != nil {
		t.Error("commit trie err", err)
	}
	state.db.TrieDB().Commit(root, false)

	cpy1, _ := New(root, tridb)
	for i := 0; i < 4; i++ {
		addr := common.BytesToAddress([]byte{byte(i)})
		for j := 0; j < 4; j++ {
			key := []byte("sk" + strconv.Itoa(i) + strconv.Itoa(j))
			value := []byte("sv1" + strconv.Itoa(i) + strconv.Itoa(j))
			cpy1.SetState(addr, common.BytesToHash(key), common.BytesToHash(value))
		}
		cpy1.IntermediateRoot(true)
	}
	root1, err := cpy1.Commit(true)
	if err != nil {
		t.Error("commit trie err", err)
	}
	cpy1.db.TrieDB().Commit(root1, false)

	cpy2, _ := New(root1, tridb)
	for i := 0; i < 4; i++ {
		addr := common.BytesToAddress([]byte{byte(i)})
		for j := 0; j < 4; j++ {
			key := []byte("sk" + strconv.Itoa(i) + strconv.Itoa(j))
			s := cpy2.GetState(addr, common.BytesToHash(key))
			fmt.Println(string(s[:]))
		}
	}
}

func TestSetAccount(t *testing.T) {
	tridb := newtridb()
	state, _ := New(common.Hash{}, tridb)
	for i := 0; i < 4; i++ {
		addr := common.BytesToAddress([]byte{byte(i)})
		for j := 0; j < 4; j++ {
			key := "at" + strconv.Itoa(int(i)) + strconv.Itoa(int(j))

			value := []byte("av" + strconv.Itoa(i) + strconv.Itoa(j))
			state.SetAccount(addr, key, value)
		}
		state.IntermediateRoot(true)
	}
	root, err := state.Commit(true)
	if err != nil {
		t.Error("commit trie err", err)
	}
	state.db.TrieDB().Commit(root, false)

	cpy1, _ := New(root, tridb)
	for i := 0; i < 4; i++ {
		addr := common.BytesToAddress([]byte{byte(i)})
		for j := 0; j < 4; j++ {
			key := "at" + strconv.Itoa(int(i)) + strconv.Itoa(int(j))
			value := []byte("av1" + strconv.Itoa(i) + strconv.Itoa(j))
			cpy1.SetAccount(addr, key, value)
		}
		cpy1.IntermediateRoot(true)
	}
	root1, err := cpy1.Commit(true)
	if err != nil {
		t.Error("commit trie err", err)
	}
	cpy1.db.TrieDB().Commit(root1, false)

	cpy2, _ := New(root1, tridb)
	for i := 0; i < 4; i++ {
		addr := common.BytesToAddress([]byte{byte(i)})
		for j := 0; j < 4; j++ {
			key := "at" + strconv.Itoa(int(i)) + strconv.Itoa(int(j))
			s := cpy2.GetAccount(addr, key)
			fmt.Println(string(s[:]))
		}
	}
}

func TestDeleteAccount(t *testing.T) {
	tridb := newtridb()
	state, _ := New(common.Hash{}, tridb)
	for i := 0; i < 4; i++ {
		addr := common.BytesToAddress([]byte{byte(i)})
		for j := 0; j < 4; j++ {
			key := "at" + strconv.Itoa(int(i)) + strconv.Itoa(int(j))
			value := []byte("av" + strconv.Itoa(i) + strconv.Itoa(j))
			state.SetAccount(addr, key, value)
		}
		state.IntermediateRoot(true)
	}
	root, err := state.Commit(true)
	if err != nil {
		t.Error("commit trie err", err)
	}
	state.db.TrieDB().Commit(root, false)

	cpy1, _ := New(root, tridb)
	for i := 0; i < 4; i++ {
		addr := common.BytesToAddress([]byte{byte(i)})
		for j := 0; j < 4; j++ {
			key := "at" + strconv.Itoa(int(i)) + strconv.Itoa(int(j))
			value := []byte("av1" + strconv.Itoa(i) + strconv.Itoa(j))
			cpy1.SetAccount(addr, key, value)
			if i == 1 {
				cpy1.DeleteAccount(addr, key)
			}
		}
		cpy1.IntermediateRoot(true)
	}
	root1, err := cpy1.Commit(true)
	if err != nil {
		t.Error("commit trie err", err)
	}
	cpy1.db.TrieDB().Commit(root1, false)

	cpy2, _ := New(root1, tridb)
	for i := 0; i < 4; i++ {
		addr := common.BytesToAddress([]byte{byte(i)})
		for j := 0; j < 4; j++ {
			key := "at" + strconv.Itoa(int(i)) + strconv.Itoa(int(j))

			s := cpy2.GetAccount(addr, key)
			fmt.Println(string(s[:]))
		}
	}
}
