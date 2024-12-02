package randomization

import (
	"base/actions/types"
	"base/config"
	"base/logger"

	"github.com/ethereum/go-ethereum/common"
)

func isValidPoolAction(actionType types.ActionType, actionsList []string) bool {
	poolName := getPoolName(actionType)
	if poolName == "" {
		return false
	}

	poolActions := getPoolActions(poolName, actionsList)

	lastPoolAction := ""
	if len(poolActions) > 0 {
		lastAction := types.ActionType(poolActions[len(poolActions)-1])
		if isDepositAction(lastAction) {
			lastPoolAction = "deposit"
		} else if isWithdrawAction(lastAction) {
			lastPoolAction = "withdraw"
		}
	}

	switch {
	case isDepositAction(actionType) && lastPoolAction == "deposit":
		logger.GlobalLogger.Infof("Действие %s нарушает чередование: последнее было deposit.", actionType)
		return false
	case isWithdrawAction(actionType) && lastPoolAction == "withdraw":
		logger.GlobalLogger.Infof("Действие %s нарушает чередование: последнее было withdraw.", actionType)
		return false
	case isWithdrawAction(actionType) && !hasCorrespondingDeposit(actionType, actionsList):
		logger.GlobalLogger.Infof("Действие %s не имеет соответствующего депозита.", actionType)
		return false
	default:
		return true
	}
}

func isDepositAction(actionType types.ActionType) bool {
	switch actionType {
	case types.AaveETHDepositAction, types.AaveUSDCSupplyAction, types.MoonwellDepositAction:
		return true
	default:
		return false
	}
}

func isLastActionDeposit(lastActions []string) bool {
	if len(lastActions) == 0 {
		return false
	}
	lastAction := types.ActionType(lastActions[len(lastActions)-1])
	return isDepositAction(lastAction)
}

func isWithdrawAction(actionType types.ActionType) bool {
	switch actionType {
	case types.AaveETHWithdrawAction, types.AaveUSDCWithdrawAction, types.MoonwellWithdrawAction:
		return true
	default:
		return false
	}
}

func getPoolActions(poolName string, actionsList []string) []string {
	poolActions := []string{}
	for _, actionStr := range actionsList {
		actionType := types.ActionType(actionStr)
		if getPoolName(actionType) == poolName {
			poolActions = append(poolActions, actionStr)
		}
	}
	return poolActions
}

func getPoolName(actionType types.ActionType) string {
	switch actionType {
	case types.AaveETHDepositAction, types.AaveETHWithdrawAction:
		return "AaveETH"
	case types.AaveUSDCSupplyAction, types.AaveUSDCWithdrawAction:
		return "AaveUSDC"
	case types.MoonwellDepositAction, types.MoonwellWithdrawAction:
		return "Moonwell"
	default:
		return ""
	}
}

func hasCorrespondingDeposit(actionType types.ActionType, actionsList []string) bool {
	expectedDepositAction := ""
	switch actionType {
	case types.AaveETHWithdrawAction:
		expectedDepositAction = string(types.AaveETHDepositAction)
	case types.AaveUSDCWithdrawAction:
		expectedDepositAction = string(types.AaveUSDCSupplyAction)
	case types.MoonwellWithdrawAction:
		expectedDepositAction = string(types.MoonwellDepositAction)
	default:
		return false
	}

	for _, action := range actionsList {
		if string(action) == expectedDepositAction {
			return true
		}
	}
	return false
}

func isValidTokenForLiquidAction(actionType types.ActionType, token common.Address) bool {
	validTokensForActions := map[types.ActionType]map[common.Address]struct{}{
		types.AaveETHDepositAction:   {config.WETH: {}},
		types.AaveUSDCSupplyAction:   {config.USDC: {}},
		types.AaveUSDCWithdrawAction: {config.USDC: {}},
		types.AaveETHWithdrawAction:  {config.WETH: {}},
		types.MoonwellDepositAction:  {config.WETH: {}},
		types.MoonwellWithdrawAction: {config.WETH: {}},
	}

	validTokens, exists := validTokensForActions[actionType]
	if !exists {
		return false
	}

	_, isValid := validTokens[token]
	return isValid
}
