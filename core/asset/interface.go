package asset

import (
	"github.com/zipper-project/z0/common"
)

// StateDB is an Asset database for full state querying.
type StateDB interface {
	GetAccount(addr common.Address, key string) []byte
	SetAccount(addr common.Address, key string, value []byte)
}
