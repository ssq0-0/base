package randomization

import (
	"base/account"
	"base/actions/types"
	"crypto/rand"
	"errors"
	"math/big"
)

func getAvailableActions(cfg account.ModulesConfig) []types.ActionType {
	actionMap := map[types.ActionType]bool{}
	if cfg.Uniswap {
		actionMap[types.UniswapAction] = true
	}
	if cfg.Pancake {
		actionMap[types.PancakeAction] = true
	}
	if cfg.Woofi {
		actionMap[types.WoofiAction] = true
	}
	if cfg.Zora {
		actionMap[types.ZoraAction] = true
	}
	if cfg.NFT2Me {
		actionMap[types.NFT2MeAction] = true
	}
	if cfg.BaseNames {
		actionMap[types.BaseNameAction] = true

	}
	if cfg.Stargate {
		actionMap[types.BridgeAction] = true
	}
	if cfg.Dmail {
		actionMap[types.DmailAction] = true
	}
	if cfg.Aave {
		actionMap[types.AaveETHDepositAction] = true
		actionMap[types.AaveETHWithdrawAction] = true
		actionMap[types.AaveUSDCSupplyAction] = true
		actionMap[types.AaveUSDCWithdrawAction] = true
	}
	if cfg.Moonwell {
		actionMap[types.MoonwellDepositAction] = true
		actionMap[types.MoonwellWithdrawAction] = true
	}
	if cfg.Collector {
		actionMap[types.CollectorModAction] = true
	}

	availableActionTypes := make([]types.ActionType, 0, len(actionMap))
	for action := range actionMap {
		availableActionTypes = append(availableActionTypes, action)
	}
	return availableActionTypes
}

func getNumActions(walletCfg account.WalletConfig) (int, error) {
	min := 15
	if walletCfg.ActionNumMIN != nil {
		min = *walletCfg.ActionNumMIN
	}

	max := 25
	if walletCfg.ActionNumMAX != nil {
		max = *walletCfg.ActionNumMAX
	}

	if min > max {
		return 0, errors.New("ActionNumMIN не может быть больше ActionNumMAX")
	}

	return generateRandomInt(min, max)
}

func generateRandomInt(min, max int) (int, error) {
	diff := max - min + 1
	randomBig, err := rand.Int(rand.Reader, big.NewInt(int64(diff)))
	if err != nil {
		return 0, err
	}
	return int(randomBig.Int64()) + min, nil
}

func updateActionHistory(lastActions []string, actionType types.ActionType) []string {
	if len(lastActions) >= 1 {
		lastActions = lastActions[1:]
	}
	lastActions = append(lastActions, string(actionType))
	return lastActions
}
