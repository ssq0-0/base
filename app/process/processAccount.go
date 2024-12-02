package process

import (
	"base/account"
	"base/actions"
	actionhandlers "base/actions/handlers"
	"base/actions/randomization"
	"base/actions/types"
	"base/app/helpers"
	cfg "base/config"
	"base/ethClient"
	"base/logger"
	"base/modules"
	"fmt"
	"math/big"
	"strings"
	"time"
)

func ProcessAccount(acc *account.Account, accConfig account.RandomConfig, mainConfig *cfg.Config, clients map[string]*ethClient.Client, randomizer *randomization.Randomizer, mods *modules.Modules, memory *Memory) {
	if strings.TrimSpace(acc.Bridge) != "" && strings.TrimSpace(acc.TokenBridge) != "" {
		if err := bridgeToBase(acc, mainConfig, clients, mods); err != nil {
			logger.GlobalLogger.Warn(err)
		}
	}
	logger.GlobalLogger.Infof("Начало обработки аккаунта %d.", acc.AccountID)

	state, err := memory.LoadState(acc.AccountID)
	if err != nil {
		logger.GlobalLogger.Errorf("Ошибка загрузки состояния для аккаунта %d: %v", acc.AccountID, err)
		return
	}

	var (
		actionSequence []actions.Action
		intervals      []time.Duration
		totalDuration  time.Duration
	)
	if state != nil && len(state.GeneratedActions) > 0 {
		lastProcessedIndex := len(state.CompletedActions)

		if lastProcessedIndex >= len(state.GeneratedActions) {
			logger.GlobalLogger.Infof("Аккаунт %d уже выполнил все действия.", acc.AccountID)
			return
		}

		actionSequence = state.GeneratedActions[lastProcessedIndex:]
		intervals = state.GeneratedIntervals[lastProcessedIndex:]

		logger.GlobalLogger.Infof("Продолжаем выполнение для аккаунта %d с действия %d", acc.AccountID, lastProcessedIndex+1)
	} else {
		actionSequence, err = randomizer.GenerateActionSequence(accConfig.Modules, accConfig.Wallets[acc.AccountID-1], acc)
		if err != nil {
			logger.GlobalLogger.Errorf("Ошибка генерации действий для аккаунта %d: %v", acc.AccountID, err)
			return
		}

		totalDuration = helpers.GetRandomDuration(acc.ActionTimeMIN, acc.ActionTimeMAX)
		intervals = helpers.DistributeActionsOverDuration(len(actionSequence), totalDuration)

		state = &AccountState{
			AccountID:          acc.AccountID,
			GeneratedActions:   actionSequence,
			GeneratedDuration:  totalDuration,
			GeneratedIntervals: intervals,
		}
		if err = memory.SaveState(state); err != nil {
			logger.GlobalLogger.Errorf("Ошибка сохранения состояния для аккаунта %d: %v", acc.AccountID, err)
		}
	}

	formattedSequence := helpers.FormatActionSequence(actionSequence, intervals)
	logger.GlobalLogger.Infof("Сгенерированная последовательность действий для аккаунта %d:\n%s", acc.AccountID, formattedSequence)

	for idx, action := range actionSequence {
		logger.GlobalLogger.Infof("Account %d waits %v before executing action %d.", acc.AccountID, intervals[idx], idx+1)
		time.Sleep(intervals[idx])
		logger.GlobalLogger.Infof("Account %d starts executing action: %s.", acc.AccountID, action.Type)

		if err = action.TakeActions(*mods, acc, action, clients["base"], mainConfig); err != nil {
			logger.GlobalLogger.Warnf("Error executing action (%s) for account %d: %v", action.Type, acc.AccountID, err)
		} else {
			logger.GlobalLogger.Infof("Action (%s) for account %d executed successfully.", action.Type, acc.AccountID)
		}

		if err := memory.UpdateState(acc.AccountID, action, intervals[idx]); err != nil {
			logger.GlobalLogger.Errorf("Error updating state for account %d: %v", acc.AccountID, err)
		}
	}

	logger.GlobalLogger.Infof("Завершение обработки аккаунта %d.", acc.AccountID)
	memory.ClearState()
}

func bridgeToBase(acc *account.Account, mainConfig *cfg.Config, clients map[string]*ethClient.Client, mods *modules.Modules) error {
	tokenAddress, exists := cfg.OtherTokens[fmt.Sprintf("%s_%s", acc.Bridge, acc.TokenBridge)]
	if !exists {
		return fmt.Errorf("токен '%s' не найден в конфигурации для аккаунта %d", acc.TokenBridge, acc.AccountID)
	}

	amountToBridge, err := clients[acc.Bridge].BalanceCheck(acc.Address, tokenAddress)
	if err != nil {
		return fmt.Errorf("failed bridge: %v", err)
	}

	reduction := new(big.Int).Div(new(big.Int).Mul(amountToBridge, big.NewInt(5)), big.NewInt(100))
	amountToBridge.Sub(amountToBridge, reduction)

	if !isNativeToken(acc.TokenBridge) {
		_, err = clients[acc.Bridge].ApproveTx(tokenAddress, cfg.LZ_Main_CA[acc.Bridge]["swap_ca"], acc.Address, acc.PrivateKey, amountToBridge, false)
		if err != nil {
			logger.GlobalLogger.Errorf("failed approve: %v", err)
		}
	}

	time.Sleep(time.Second * 5)
	if err := actionhandlers.BridgeHandler.Execute(actionhandlers.BridgeHandler{
		BridgeParams: types.BridgeParams{
			FromChain:      acc.Bridge,
			DstChain:       cfg.LZ_Chain_ids["Base"],
			SrcPoolId:      cfg.LZ_Pool_ids[fmt.Sprintf("%s_%s", acc.Bridge, acc.TokenBridge)],
			DstPoolId:      cfg.LZ_Pool_ids["base_usdc"],
			AmountToBridge: amountToBridge,
		},
	}, acc, *mods, clients[acc.Bridge], mainConfig); err != nil {
		logger.GlobalLogger.Warnf("bridge error: %v", err)
	}
	if acc.Bridge == "polygon" {
		logger.GlobalLogger.Info("бридж окончен, спим 25 минут и ждем поступления токенов в сеть. При необходимости перезапустить без модуля bridge")
		time.Sleep(time.Second * 1500)
	}
	logger.GlobalLogger.Info("бридж окончен, спим 3 минуты и ждем поступления токенов в сеть. При необходимости перезапустить без модуля bridge")
	time.Sleep(time.Second * 180)

	return nil
}

func isNativeToken(tokenSymbol string) bool {
	nativeTokens := []string{"ETH"}
	for _, nativeToken := range nativeTokens {
		if strings.EqualFold(tokenSymbol, nativeToken) {
			return true
		}
	}
	return false
}
