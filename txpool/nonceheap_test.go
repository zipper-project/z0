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

package txpool

import (
	"sort"
	"testing"

	"github.com/zipper-project/z0/common"
)

func TestNonceHeap(t *testing.T) {
	var nh nonceHeap

	array := []uint64{2, 1, 4, 3}
	for _, v := range array {
		nh.Push(v)
	}
	for i := 0; i < 4; i++ {
		common.AssertEquals(t, array[3-i], nh.Pop().(uint64))
	}

	//test sort
	sortarray := []uint64{4, 3, 2, 1}

	for _, v := range array {
		nh.Push(v)
	}
	sort.Sort(nh)
	for i := 0; i < 4; i++ {
		common.AssertEquals(t, sortarray[i], nh.Pop().(uint64))
	}
}
