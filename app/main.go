package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"base/account"
	"base/ethClient"
	"base/logger"
	"base/utils"

	"base/actions/randomization"
	"base/app/helpers"
	"base/app/process"
	cfg "base/config"
	"base/modules"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	logger.GlobalLogger.Info(cfg.Logo)
	logger.GlobalLogger.Info(cfg.Subscribe)
	logger.GlobalLogger.Infof(cfg.DonateSOL)
	logger.GlobalLogger.Infof(cfg.DonateEVM)
	logger.GlobalLogger.Infof(cfg.DonateEVM)
	time.Sleep(time.Second * 5)

	accConfig, err := account.LoadRandomConfig(fmt.Sprintf("%s/%s", utils.GetRootDir(), "account/account_config.json"))
	if err != nil {
		logger.GlobalLogger.Fatalf("не удалось загрузить рандомную конфигурацию: %v", err)
	}
	logger.GlobalLogger.Info("конфигурация рандомизации успешно загружена.")

	accounts, err := account.CreateAccounts(accConfig.Wallets)
	if err != nil {
		logger.GlobalLogger.Fatalf("ошибка создания аккаунтов: %v", err)
	}

	if len(accounts) == 0 {
		logger.GlobalLogger.Fatalf("Нет созданных аккаунтов. Проверьте приватные ключи в конфигурации.")
	}

	config, err := cfg.LoadConfig(fmt.Sprintf("%s/%s", utils.GetRootDir(), "config/config.json"))
	if err != nil {
		logger.GlobalLogger.Fatalf("ошибка загрузки основного конфига, проверьте его целостность: %v", err)
	}
	logger.GlobalLogger.Info("Основная конфигурация успешно загружена.")

	var clients = make(map[string]*ethClient.Client)
	for chain, rpc := range cfg.RPCs {
		client := ethClient.NewClient(rpc)
		if client == nil {
			logger.GlobalLogger.Errorf("ошибка создания eth client. Проверьте RPC.")
			continue
		}
		clients[chain] = client
	}
	defer ethClient.CloseAllClients(clients)

	mods, err := modules.InitializeModules(*config, clients)
	if err != nil {
		logger.GlobalLogger.Fatalf("ошибка инициализации модулей: %v", err)
	}
	logger.GlobalLogger.Info("Все модули успешно инициализированы. Спим 2 секунды.")
	time.Sleep(time.Second * 2)

	stateFilePath := fmt.Sprintf("%s/%s", utils.GetRootDir(), "app/process/state.json")
	memoryHandler := process.NewMemory(stateFilePath)
	stateExists, err := memoryHandler.IsStateFileNotEmpty()
	if err != nil {
		logger.GlobalLogger.Errorf("Ошибка проверки состояния: %v", err)
		return
	}

	if stateExists {
		var userInput string
		fmt.Println("Продолжить выполнение? (y/n): ")
		fmt.Scanln(&userInput)

		if userInput != "y" {
			err := memoryHandler.ClearState()
			if err != nil {
				logger.GlobalLogger.Errorf("Ошибка очистки состояния: %v", err)
				return
			}
		}
	} else {
		logger.GlobalLogger.Info("Файл состояния пуст. Начинаем выполнение с чистого листа.")
	}

	availableNFTs, aviableTokens := account.InitializeAvailableNFTs(accConfig), account.ConvertStringsToAddresses(accConfig.Tokens)
	randomizer := randomization.NewRandomizer(helpers.AvailableTokensToSlice(aviableTokens), availableNFTs, clients)

	var wg sync.WaitGroup
	for _, acc := range accounts {
		wg.Add(1)
		go func(acc *account.Account) {
			defer wg.Done()
			process.ProcessAccount(acc, accConfig, config, clients, randomizer, mods, memoryHandler)
		}(acc)
	}

	wg.Wait()
	logger.GlobalLogger.Infof("Все действия выполнены. Программа завершает работу.")
}
