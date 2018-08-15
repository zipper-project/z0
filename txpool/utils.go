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
	"math"
	"reflect"

	"github.com/zipper-project/z0/common"
	"github.com/zipper-project/z0/params"
	"github.com/zipper-project/z0/types"
)

// TxDifference returns a new set which is the difference between a and b.
func TxDifference(a, b types.Transactions) types.Transactions {
	keep := make(types.Transactions, 0, len(a))

	remove := make(map[common.Hash]struct{})
	for _, tx := range b {
		remove[tx.Hash()] = struct{}{}
	}

	for _, tx := range a {
		if _, ok := remove[tx.Hash()]; !ok {
			keep = append(keep, tx)
		}
	}

	return keep
}

// IntrinsicGas computes the 'intrinsic gas' for a message with the given data.
func IntrinsicGas(Extra []byte, inputs, outputs []interface{}) (uint64, error) {
	var gas uint64
	for _, v := range outputs {
		if reflect.TypeOf(v) == types.AMOutputType {
			output := v.(types.AMOutput)
			if output.Address == nil {
				gas += params.TxGasContractCreation
			} else {
				gas += params.TxGas
			}
		}
	}

	// Bump the required gas by the amount of transactional data
	dataGasFunc := func(data []byte) (uint64, error) {
		var gas uint64
		if len(data) > 0 {
			// Zero and non-zero bytes are priced differently
			var nz uint64
			for _, byt := range data {
				if byt != 0 {
					nz++
				}
			}
			// Make sure we don't exceed uint64 for all data combinations
			if (math.MaxUint64-gas)/params.TxDataNonZeroGas < nz {
				return 0, ErrOutOfGas
			}
			gas += nz * params.TxDataNonZeroGas

			z := uint64(len(data)) - nz
			if (math.MaxUint64-gas)/params.TxDataZeroGas < z {
				return 0, ErrOutOfGas
			}
			gas += z * params.TxDataZeroGas
		}
		return gas, nil
	}

	for _, v := range inputs {
		if reflect.TypeOf(v) == types.AMInputType {
			input := v.(types.AMInput)
			dataGas, err := dataGasFunc(input.Payload)
			if err != nil {
				return 0, err
			}
			gas += dataGas
		}
	}

	return gas, nil
}
