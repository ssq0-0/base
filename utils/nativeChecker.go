package utils

import (
	"base/config"

	"github.com/ethereum/go-ethereum/common"
)

func IsNativeToken(tokenAddr common.Address) bool {
	nativeTokens := map[common.Address]bool{
		config.WETH:     true,
		config.WooFiETH: true,
	}

	return nativeTokens[tokenAddr]
}
