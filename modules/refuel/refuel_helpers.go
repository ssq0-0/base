package refuel

import (
	"errors"
	"math/big"
	"math/rand"
	"time"
)

func (rf *Refuel) getBridgeAmount(chain string) (*big.Int, error) {
	var min, max *big.Int

	switch chain {
	case "polygon":
		min = big.NewInt(4000000000000000000)
		max = big.NewInt(4500000000000000000)
	case "avalanche":
		min = big.NewInt(87659812000000000)
		max = big.NewInt(125439870000000000)
	case "arbitrum", "optimism", "base":
		min = big.NewInt(789000000000000)
		max = big.NewInt(1230000000000000)
	default:
		return nil, errors.New("неизвестная сеть для бриджа: " + chain)
	}

	diff := new(big.Int).Sub(max, min)
	randomOffset := new(big.Int).Rand(rand.New(rand.NewSource(time.Now().UnixNano())), diff)
	return new(big.Int).Add(min, randomOffset), nil
}
