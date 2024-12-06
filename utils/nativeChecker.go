package utils

import (
	"base/config"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

func IsNativeToken(tokenAddr common.Address) bool {
	nativeTokens := map[common.Address]bool{
		config.WETH:     true,
		config.WooFiETH: true,
	}

	return nativeTokens[tokenAddr]
}

func IsNativeTokenBySymbol(tokenSymbol string) bool {
	nativeTokens := []string{"ETH", "AVAX", "MATIC"}
	for _, nativeToken := range nativeTokens {
		if strings.EqualFold(tokenSymbol, nativeToken) {
			return true
		}
	}
	return false
}

func CheckAdaptDecimals(tokenSymbol string) {

}
