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

type MoonwellHandler struct {
	LiquidParams types.LiquidParams
}

func (mh MoonwellHandler) Execute(acc *account.Account, mods modules.Modules, client *ethClient.Client, config *cfg.Config) error {
	amount, err := mh.CalculateAmountToDeposit(acc, client, cfg.WETH)
	if err != nil {
		return err
	}

	switch mh.LiquidParams.Type {
	case string(types.MoonwellDepositAction):
		return mods.LiquidPools.Moonwell.DepositETH(amount, acc)
	case string(types.MoonwellWithdrawAction):
		return mods.LiquidPools.Moonwell.WithdrawETH(acc, cfg.WETH)
	default:
		return fmt.Errorf("unknown action type: %s", mh.LiquidParams.Type)
	}
}

func (mh MoonwellHandler) CalculateAmountToDeposit(acc *account.Account, client *ethClient.Client, token common.Address) (*big.Int, error) {
	return CalculatePercentageOfBalance(acc, client, token, acc.PoolUsedRange, []common.Address{cfg.WETH})
}
