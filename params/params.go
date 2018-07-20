// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package params

const (
	// TxGas Per transaction not creating a contract. NOTE: Not payable on data of calls between transactions.
	TxGas uint64 = 21000
	// TxGasContractCreation Per transaction that creates a contract. NOTE: Not payable on data of calls between transactions.
	TxGasContractCreation uint64 = 53000
	// TxDataNonZeroGas Per byte of data attached to a transaction that is not equal to zero. NOTE:Not payable on data of calls between transactions.
	TxDataNonZeroGas uint64 = 68
	// TxDataZeroGas Per byte of data attached to a transaction that equals zero. NOTE: Not payable on data of calls between transactions.
	TxDataZeroGas uint64 = 4
)
