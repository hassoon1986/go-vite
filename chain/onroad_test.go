package chain

import (
	"fmt"
	"github.com/vitelabs/go-vite/common/types"
	"testing"
)

func TestChain_OnRoad(t *testing.T) {
	chainInstance, accounts, _ := SetUp(t, 123, 1231, 12)

	testOnRoad(t, chainInstance, accounts)

	TearDown(chainInstance)
}

func testOnRoad(t *testing.T, chainInstance *chain, accounts map[types.Address]*Account) {
	t.Run("HasOnRoadBlocks", func(t *testing.T) {
		HasOnRoadBlocks(t, chainInstance, accounts)
	})

	t.Run("GetOnRoadBlocksHashList", func(t *testing.T) {
		GetOnRoadBlocksHashList(t, chainInstance, accounts)
	})
}

func HasOnRoadBlocks(t *testing.T, chainInstance *chain, accounts map[types.Address]*Account) {
	for addr, account := range accounts {
		result, err := chainInstance.HasOnRoadBlocks(addr)
		if err != nil {
			t.Fatal(err)
		}

		if result && len(account.UnreceivedBlocks) <= 0 {
			t.Fatal(fmt.Sprintf("%s", addr))
		}

		if !result && len(account.UnreceivedBlocks) > 0 {
			t.Fatal("error")
		}
	}
}

func GetOnRoadBlocksHashList(t *testing.T, chainInstance Chain, accounts map[types.Address]*Account) {
	countPerPage := 10

	for addr, account := range accounts {
		pageNum := 0
		hashSet := make(map[types.Hash]struct{})

		for {
			hashList, err := chainInstance.GetOnRoadBlocksHashList(addr, pageNum, 10)
			if err != nil {
				t.Fatal(err)
			}

			hashListLen := len(hashList)
			if hashListLen <= 0 {
				break
			}

			if hashListLen > countPerPage {
				t.Fatal(err)
			}

			for _, hash := range hashList {
				if _, ok := hashSet[hash]; ok {
					t.Fatal(err)
				}

				hashSet[hash] = struct{}{}

				if _, hasUnReceive := account.UnreceivedBlocks[hash]; !hasUnReceive {
					t.Fatal("error")
				}
			}
			pageNum++
		}

		if len(hashSet) != len(account.UnreceivedBlocks) {
			t.Fatal("error")
		}
	}
}
