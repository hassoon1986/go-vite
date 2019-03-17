package pmchain

import (
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/ledger"
	"math/big"
)

func (c *chain) GetBalance(addr *types.Address, tokenId *types.TokenTypeId) (*big.Int, error) {
	return nil, nil
}

// get history balance, if history is too old, failed
func (c *chain) GetHistoryBalance(accountBlockHash *types.Hash, addr *types.Address, tokenId *types.TokenTypeId) (*big.Int, error) {
	return nil, nil
}

// get confirmed snapshot balance, if history is too old, failed
func (c *chain) GetConfirmedBalance(addr *types.Address, tokenId *types.TokenTypeId, sbHash *types.Hash) (*big.Int, error) {
	return nil, nil
}

// get contract code
func (c *chain) GetContractCode(contractAddr *types.Address) ([]byte, error) {
	return nil, nil
}

func (c *chain) GetContractMeta(contractAddress *types.Address) (meta *ledger.ContractMeta, err error) {
	return nil, nil
}

func (c *chain) GetContractList(gid *types.Gid) (map[types.Address]*ledger.ContractMeta, error) {
	return nil, nil
}

func (c *chain) GetQuotaUnused(address *types.Address) uint64 {
	return 0
}

func (c *chain) GetQuotaUsed(address *types.Address) (quotaUsed uint64, blockCount uint64) {
	return 0, 0
}
