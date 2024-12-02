package handlers

import (
	"base/account"
	"base/actions/types"
	cfg "base/config"
	"base/ethClient"
	"base/modules"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type DexHandler struct {
	DexParams  types.DexParams
	ActionType types.ActionType
}

func (dh DexHandler) Execute(acc *account.Account, mods modules.Modules, client *ethClient.Client, config *cfg.Config) error {
	amountToSwap, err := dh.calculateAmountToSwap(acc, client)
	if err != nil {
		return err
	}
	if amountToSwap.Cmp(big.NewInt(0)) == 0 {
		return errors.New("invalid amount to swap")
	}

	var value *big.Int
	if dh.isNativeToken(dh.DexParams.FromToken) {
		value = amountToSwap
	}

	switch dh.ActionType {
	case types.UniswapAction:
		dex := mods.Dex.Uniswap
		if err := dh.ensureApproval(client, acc, mods.Dex.Uniswap.RouterCA, amountToSwap); err != nil {
			return err
		}

		if dh.isNativeToken(dh.DexParams.ToToken) {
			err = dex.SwapToETH(dh.DexParams.FromToken, dh.DexParams.ToToken, amountToSwap, big.NewInt(0), acc)
		} else {
			err = dex.Swap(dh.DexParams.FromToken, dh.DexParams.ToToken, amountToSwap, value, acc)
		}
	case types.PancakeAction:
		dex := mods.Dex.Pancake
		if err := dh.ensureApproval(client, acc, mods.Dex.Pancake.RouterCA, amountToSwap); err != nil {
			return err
		}

		if dh.isNativeToken(dh.DexParams.ToToken) {
			err = dex.SwapToETH(dh.DexParams.FromToken, dh.DexParams.ToToken, amountToSwap, big.NewInt(0), acc)
		} else {
			err = dex.Swap(dh.DexParams.FromToken, dh.DexParams.ToToken, amountToSwap, value, acc)
		}
	case types.WoofiAction:
		dex := mods.Dex.Woofi
		if dh.DexParams.FromToken == cfg.WETH {
			dh.DexParams.FromToken = cfg.WooFiETH
		}
		if dh.DexParams.ToToken == cfg.WETH {
			dh.DexParams.ToToken = cfg.WooFiETH
		}

		if err := dh.ensureApproval(client, acc, dex.CA, amountToSwap); err != nil {
			return err
		}

		err = dex.Swap(dh.DexParams.FromToken, dh.DexParams.ToToken, amountToSwap, value, acc)
	default:
		return errors.New("unsupported DEX action type")
	}

	return err
}

func (dh *DexHandler) ensureApproval(client *ethClient.Client, acc *account.Account, routerCA common.Address, value *big.Int) error {
	_, err := client.ApproveTx(dh.DexParams.FromToken, routerCA, acc.Address, acc.PrivateKey, value, false)
	return err
}

func (dh *DexHandler) isNativeToken(token common.Address) bool {
	return token == cfg.WETH || token == cfg.WooFiETH
}

func (dh *DexHandler) calculateAmountToSwap(acc *account.Account, client *ethClient.Client) (*big.Int, error) {
	return CalculatePercentageOfBalance(acc, client, dh.DexParams.FromToken, acc.UsedRange, []common.Address{cfg.WETH, cfg.WooFiETH})
}
