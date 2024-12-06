package helpers

import (
	"base/account"
	"base/config"
	"base/ethClient"
	"base/logger"
	"base/utils"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func PrintStartupMessages() {
	logger.GlobalLogger.Info(config.Logo)
	logger.GlobalLogger.Info(config.Subscribe)
	logger.GlobalLogger.Infof(config.DonateSOL)
	logger.GlobalLogger.Infof(config.DonateEVM)
	time.Sleep(5 * time.Second)
}

func AllPathInit() (string, string, string, error) {
	rootDir := utils.GetRootDir()
	log.Printf("rootDir: %s", rootDir)
	accConfigPath := filepath.Join(rootDir, "account", "account_config.json")
	if _, err := os.Stat(accConfigPath); os.IsNotExist(err) {
		return "", "", "", fmt.Errorf("файл не найден: %s", accConfigPath)
	}
	log.Printf("accConfigPath: %s", accConfigPath)

	configPath := filepath.Join(rootDir, "config", "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", "", "", fmt.Errorf("файл не найден: %s", configPath)
	}
	log.Printf("configPath: %s", configPath)

	stateFilePath := filepath.Join(rootDir, "app", "process", "state.json")
	if _, err := os.Stat(stateFilePath); os.IsNotExist(err) {
		file, err := os.Create(stateFilePath)
		if err != nil {
			return "", "", "", fmt.Errorf("не удалось создать файл состояния: %v", err)
		}
		defer file.Close()
	}
	log.Printf("statepath: %s", stateFilePath)

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
		client, err := ethClient.NewClient(rpc)
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
