package bridge

import (
	"base/ethClient"
	"base/models"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

func getFee(client *ethClient.Client, toCA common.Address, abi *abi.ABI, ownerAddr common.Address, dstChain uint16, methodName string) (*big.Int, error) {
	feeData, err := abi.Pack("quoteLayerZeroFee", dstChain, uint8(1), ownerAddr.Bytes(), []byte{}, models.LzTxObj{
		DstGasForCall:   big.NewInt(0),
		DstNativeAmount: big.NewInt(0),
		DstNativeAddr:   common.Hex2Bytes("0000000000000000000000000000000000000001"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed pack data for quoteLayerZeroFee: %v", err)
	}

	feeResult, err := client.CallCA(toCA, feeData)
	if err != nil {
		return nil, fmt.Errorf("failed call quoteLayerZeroFee: %v", err)
	}

	feeOutputs, err := abi.Unpack(methodName, feeResult)
	if err != nil {
		return nil, fmt.Errorf("failed unpack quoteLayerZeroFee: %v", err)
	}

	if len(feeOutputs) != 2 {
		return nil, fmt.Errorf("wait 2 params forquoteLayerZeroFee, received: %d", len(feeOutputs))
	}

	feeWei, ok := feeOutputs[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("result not *big.Int")
	}
	return feeWei, nil
}

func calculateMinAmountLD(amountIn *big.Int) *big.Int {
	minAmount := new(big.Int).Mul(amountIn, big.NewInt(98))
	return minAmount.Div(minAmount, big.NewInt(100))
}
