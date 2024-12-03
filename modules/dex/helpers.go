package dex

import (
	"base/ethClient"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

func getAmountMin(toCA common.Address, data []byte, client *ethClient.Client, abi *abi.ABI, methodName string, slippage *big.Float) (*big.Int, error) {
	result, err := client.CallCA(toCA, data)
	if err != nil {
		return nil, err
	}

	unpackedData, err := abi.Unpack(methodName, result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack the result of tryQuerySwap: %v", err)
	}

	if len(unpackedData) < 1 {
		return nil, fmt.Errorf("empty result from tryQuerySwap")
	}

	amountOut, ok := unpackedData[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("error of conversion to *big.Int")
	}

	return applySlippage(amountOut, slippage), nil
}

func applySlippage(amount *big.Int, slippage *big.Float) *big.Int {
	amountFloat := new(big.Float).SetInt(amount)
	adjustedAmountFloat := new(big.Float).Mul(amountFloat, slippage)
	adjustedAmount, _ := adjustedAmountFloat.Int(nil)
	return adjustedAmount
}
