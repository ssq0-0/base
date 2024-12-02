package randomization

import (
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
)

func getRandomNFT(nftMap map[common.Address]*big.Int) (common.Address, *big.Int) {
	keys := make([]common.Address, 0, len(nftMap))
	for k := range nftMap {
		keys = append(keys, k)
	}
	randIndex := rand.Intn(len(keys))
	selectedContract := keys[randIndex]
	price := nftMap[selectedContract]
	return selectedContract, price
}
