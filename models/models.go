package models

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type DSNDomain struct {
	Name          string
	Owner         common.Address
	Duration      *big.Int
	Resolver      common.Address
	Data          [][]byte
	ReverseRecord bool
}

type OtherDexParams struct {
	TokenIn           common.Address
	TokenOut          common.Address
	AmountIn          *big.Int
	Fee               *big.Int
	SqrtPriceLimitX96 *big.Int
}

type ExactInputSingleParams struct {
	TokenIn           common.Address
	TokenOut          common.Address
	Fee               *big.Int
	Recipient         common.Address
	AmountIn          *big.Int
	AmountOutMinimum  *big.Int
	SqrtPriceLimitX96 *big.Int
}

type LzTxObj struct {
	DstGasForCall   *big.Int `abi:"dstGasForCall"`
	DstNativeAmount *big.Int `abi:"dstNativeAmount"`
	DstNativeAddr   []byte   `abi:"dstNativeAddr"`
}

type SwapPair struct {
	From common.Address
	To   common.Address
}
