package handlers

import (
	"base/account"
	"base/actions/types"
	cfg "base/config"
	"base/ethClient"
	"base/modules"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type AaveHandler struct {
	LiquidParams types.LiquidParams
}

func (ah AaveHandler) Execute(acc *account.Account, mods modules.Modules, client *ethClient.Client, config *cfg.Config) error {
	switch ah.LiquidParams.Type {
	case string(types.AaveETHDepositAction):
		return ah.handleDeposit(acc, mods, client)
	case string(types.AaveETHWithdrawAction):
		return ah.handleWithdrawETH(acc, client, mods)
	case string(types.AaveUSDCSupplyAction):
		return ah.handleSupply(acc, mods, client)
	case string(types.AaveUSDCWithdrawAction):
		return ah.handleWithdrawSpecific(acc, client, mods)
	default:
		return fmt.Errorf("неизвестный тип действия: %s", ah.LiquidParams.Type)
	}
}

func (ah AaveHandler) handleDeposit(acc *account.Account, mods modules.Modules, client *ethClient.Client) error {
	amount, err := ah.calculateAmountToDeposit(acc, client, cfg.WETH)
	if err != nil {
		return err
	}

	return mods.LiquidPools.Aave.DepositETH(amount, acc)
}

func (ah AaveHandler) handleWithdrawETH(acc *account.Account, client *ethClient.Client, mods modules.Modules) error {
	amount, err := client.BalanceCheck(acc.Address, cfg.AaveWETH)
	if err != nil {
		return err
	}

	if err := ah.ensureApproval(client, acc, cfg.AaveWETH, mods.LiquidPools.Aave.EthPool, amount); err != nil {
		return err
	}

	return mods.LiquidPools.Aave.WithdrawETH(acc, amount)
}

func (ah AaveHandler) handleSupply(acc *account.Account, mods modules.Modules, client *ethClient.Client) error {
	amount, err := ah.calculateAmountToDeposit(acc, client, cfg.USDC)
	if err != nil {
		return err
	}

	if err := ah.ensureApproval(client, acc, cfg.USDC, mods.LiquidPools.Aave.ProxyBase, amount); err != nil {
		return err
	}

	return mods.LiquidPools.Aave.Supply(acc, cfg.USDC, amount)
}

func (ah AaveHandler) handleWithdrawSpecific(acc *account.Account, client *ethClient.Client, mods modules.Modules) error {
	amount, err := client.BalanceCheck(acc.Address, cfg.AaveUSDC)
	if err != nil {
		return err
	}

	if err := ah.ensureApproval(client, acc, cfg.AaveUSDC, mods.LiquidPools.Aave.ProxyBase, amount); err != nil {
		return err
	}

	return mods.LiquidPools.Aave.Withdraw(acc, cfg.USDC)
}

func (ah AaveHandler) calculateAmountToDeposit(acc *account.Account, client *ethClient.Client, token common.Address) (*big.Int, error) {
	return CalculatePercentageOfBalance(acc, client, token, acc.PoolUsedRange, []common.Address{cfg.WETH})
}

func (ah AaveHandler) ensureApproval(client *ethClient.Client, acc *account.Account, tokenAddr, spender common.Address, amount *big.Int) error {
	_, err := client.ApproveTx(tokenAddr, spender, acc.Address, acc.PrivateKey, amount, false)
	return err
}
