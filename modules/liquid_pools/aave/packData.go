package aave

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

func (a *Aave) packDeposit(ownerAddr common.Address) ([]byte, error) {
	return a.ABI.Pack("depositETH", a.ProxyBase, ownerAddr, uint16(0))
}

func (a *Aave) packWithdraw(ownerAddr common.Address, amountOut *big.Int) ([]byte, error) {
	return a.ABI.Pack("withdrawETH", a.ProxyBase, amountOut, ownerAddr)
}

func (a *Aave) packSupply(tokenIn, ownerAddr common.Address, amountIn *big.Int) ([]byte, error) {
	return a.ABI.Pack("supply", tokenIn, amountIn, ownerAddr, uint16(0))
}

func (a *Aave) packWithdrawStable(tokenOut, ownerAddr common.Address, amountOut *big.Int) ([]byte, error) {
	return a.ABI.Pack("withdraw", tokenOut, amountOut, ownerAddr)
}
