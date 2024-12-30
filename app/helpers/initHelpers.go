package helpers

import (
	"base/account"
	"base/config"
	"base/ethClient"
	"base/logger"
	"errors"
	"time"
)

func PrintStartupMessages() {
	logger.GlobalLogger.Info(config.Logo)
	time.Sleep(5 * time.Second)
}

func AllPathInit() (string, string, string, error) {
	accConfigPath := "account/account_config.json"
	configPath := "config/config.json"
	stateFilePath := "app/process/state.json"

	return accConfigPath, configPath, stateFilePath, nil
}

func AccsInit(accConfigPath string) ([]*account.Account, *account.RandomConfig, error) {
	accConfig, err := account.LoadRandomConfig(accConfigPath)
	if err != nil {
		logger.GlobalLogger.Fatalf("не удалось загрузить конфигурацию для рандомизации: %v", err)
	}
	logger.GlobalLogger.Info("конфигурация рандомизации успешно загружена.")

	accounts, err := account.CreateAccounts(accConfig.Wallets)
	if err != nil {
		return nil, nil, err
	}

	return accounts, accConfig, nil
}

func ClientsInit() (map[string]*ethClient.Client, error) {
	var clients = make(map[string]*ethClient.Client)
	for chain, rpc := range config.RPCs {
		client, err := ethClient.NewClient(rpc, "account/account_stats.txt")
		if err != nil {
			logger.GlobalLogger.Errorf("Ошибка создания eth client для сети %s: %v", chain, err)
			continue
		}
		clients[chain] = client
	}

	if len(clients) == 0 {
		return nil, errors.New("не удалось создать ни одного клиента. Проверьте настройки RPC")
	}

	return clients, nil
}
