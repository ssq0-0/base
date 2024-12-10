package process

import (
	"base/account"
	"base/actions/randomization"
	"base/app/helpers"
	"base/config"
	"base/ethClient"
	"base/logger"
	"base/modules"
	"fmt"
	"strings"
	"time"
)

func ProcessAccount(acc *account.Account, accConfig *account.RandomConfig, mainConfig *config.Config, clients map[string]*ethClient.Client, randomizer *randomization.Randomizer, mods *modules.Modules, memory *Memory) {
	if shouldBridge(acc) {
		if err := bridgeToBase(acc, mainConfig, clients, mods); err != nil {
			logger.GlobalLogger.Warn(err)
		}
	}
	logger.GlobalLogger.Infof("Начало обработки аккаунта %d.", acc.AccountID)

	state, err := loadOrCreateState(acc, accConfig, randomizer, memory)
	if err != nil {
		logger.GlobalLogger.Errorf("Ошибка с состоянием: %v", err)
		return
	}

	logger.GlobalLogger.Infof("Сгенерированная последовательность действий для аккаунта %d:\n%s",
		acc.AccountID, helpers.FormatActionSequence(state.GeneratedActions, state.GeneratedIntervals))

	executeActions(acc, state, mods, clients["base"], mainConfig, memory)
	logger.GlobalLogger.Infof("Завершение обработки аккаунта %d.", acc.AccountID)
}

func shouldBridge(acc *account.Account) bool {
	return strings.TrimSpace(acc.Bridge) != "" && strings.TrimSpace(acc.TokenBridge) != ""
}

func loadOrCreateState(acc *account.Account, accConfig *account.RandomConfig, randomizer *randomization.Randomizer, memory *Memory) (*AccountState, error) {
	state, err := memory.LoadState(acc.AccountID)
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки состояния для аккаунта %d: %w", acc.AccountID, err)
	}

	if state != nil && len(state.GeneratedActions) > 0 {
		lastProcessedIndex := len(state.CompletedActions)
		if lastProcessedIndex >= len(state.GeneratedActions) {
			logger.GlobalLogger.Infof("Аккаунт %d уже выполнил все действия.", acc.AccountID)
			return state, nil
		}
		logger.GlobalLogger.Infof("Продолжаем выполнение для аккаунта %d с действия %d", acc.AccountID, lastProcessedIndex+1)
		return state, nil
	}

	actionSequence, err := randomizer.GenerateActionSequence(&accConfig.Modules, &accConfig.Wallets[acc.AccountID-1], acc)
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации действий для аккаунта %d: %w", acc.AccountID, err)
	}

	totalDuration := helpers.GetRandomDuration(acc.ActionTimeMIN, acc.ActionTimeMAX)
	intervals := helpers.DistributeActionsOverDuration(len(actionSequence), totalDuration)

	state = &AccountState{
		AccountID:          acc.AccountID,
		GeneratedActions:   actionSequence,
		GeneratedDuration:  totalDuration,
		GeneratedIntervals: intervals,
	}

	if err = memory.SaveState(state); err != nil {
		logger.GlobalLogger.Errorf("Ошибка сохранения состояния для аккаунта %d: %v", acc.AccountID, err)
	}

	return state, nil
}

func executeActions(acc *account.Account, state *AccountState, mods *modules.Modules, client *ethClient.Client, mainConfig *config.Config, memory *Memory) {
	actionsLeft := state.GeneratedActions[len(state.CompletedActions):]
	intervalsLeft := state.GeneratedIntervals[len(state.CompletedActions):]

	for i, action := range actionsLeft {
		currentStepNumber := len(state.CompletedActions) + i + 1

		logger.GlobalLogger.Infof("Аккаунт %d ждет %v перед началом действия %d.", acc.AccountID, intervalsLeft[i], currentStepNumber)
		time.Sleep(intervalsLeft[i])

		logger.GlobalLogger.Infof("Аккаунт %d начинает действие: %s.", acc.AccountID, action.Type)
		if err := action.TakeActions(*mods, acc, action, client, mainConfig); err != nil {
			logger.GlobalLogger.Warnf("Ошибка выполнения (%s) для аккаунта %d: %v", action.Type, acc.AccountID, err)
			if strings.Contains(err.Error(), "insufficient funds") {
				if errRefuel := checkRefuel(acc, map[string]*ethClient.Client{"base": client}, mods); errRefuel != nil {
					return
				}
			}
		} else {
			logger.GlobalLogger.Infof("Действие (%s) для аккаунта %d выполнено успешно.", action.Type, acc.AccountID)
		}

		if err := memory.UpdateState(acc.AccountID, action, intervalsLeft[i]); err != nil {
			logger.GlobalLogger.Errorf("Ошибка обновления состояния аккаунта %d: %v", acc.AccountID, err)
		}
	}

	if err := memory.ClearState(acc.AccountID); err != nil {
		logger.GlobalLogger.Errorf("Ошибка очистки состояния для аккаунта %d: %v", acc.AccountID, err)
	}
	logger.GlobalLogger.Infof("Анализ аккаунтов...")
	// client.Analizor()
}
