package aave

import (
	"base/account"
	"base/config"
	"base/ethClient"
	"base/utils"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type Aave struct {
	ProxyBase common.Address
	EthPool   common.Address
	ABI       *abi.ABI
	Client    *ethClient.Client
}

func NewAave(client *ethClient.Client, proxy_base, ethPoolCa, abiPath string) (*Aave, error) {
	abi, err := utils.ReadAbi(abiPath)
	if err != nil {
		return nil, err
	}

	return &Aave{
		ProxyBase: common.HexToAddress(proxy_base),
		EthPool:   common.HexToAddress(ethPoolCa),
		Client:    client,
		ABI:       abi,
	}, nil
}

func (a *Aave) DepositETH(amountIn *big.Int, acc *account.Account) error {
	data, err := a.packDeposit(acc.Address)
	if err != nil {
		return err
	}

	return a.Client.SendTransaction(acc.PrivateKey, acc.Address, a.EthPool, a.Client.GetNonce(acc.Address), amountIn, data)
}

func (a *Aave) Supply(acc *account.Account, tokenIn common.Address, amountIn *big.Int) error {
	data, err := a.packSupply(tokenIn, acc.Address, amountIn)
	if err != nil {
		return err
	}

	return a.Client.SendTransaction(acc.PrivateKey, acc.Address, a.ProxyBase, a.Client.GetNonce(acc.Address), big.NewInt(0), data)
}

func (a *Aave) WithdrawETH(acc *account.Account, amount *big.Int) error {
	data, err := a.packWithdraw(acc.Address, amount)
	if err != nil {
		return fmt.Errorf("error packing withdrawETH: %v", err)
	}

	return a.Client.SendTransaction(acc.PrivateKey, acc.Address, a.EthPool, a.Client.GetNonce(acc.Address), big.NewInt(0), data)
}

func (a *Aave) Withdraw(acc *account.Account, tokenOut common.Address) error {
	balance, err := a.Client.BalanceCheck(acc.Address, config.AaveUSDC)
	if err != nil {
		return err
	}
	data, err := a.packWithdrawStable(tokenOut, acc.Address, balance)
	if err != nil {
		return err
	}

	return a.Client.SendTransaction(acc.PrivateKey, acc.Address, a.ProxyBase, a.Client.GetNonce(acc.Address), big.NewInt(0), data)
}

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
