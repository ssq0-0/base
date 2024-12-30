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

type AssembleResponse struct {
	Transaction Transaction `json:"transaction"`
}

type Transaction struct {
	Value string `json:"value"`
	To    string `json:"to"`
	From  string `json:"from"`
	Data  string `json:"data"`
}

type SwapQuoteResponse struct {
	Data struct {
		To           string `json:"to"`
		Data         string `json:"data"`
		Value        string `json:"value"`
		EstimatedGas int    `json:"estimatedGas"`
	} `json:"data"`
}
