package main

import (
	"base/account"
	"base/actions/randomization"
	cfg "base/config"
	"base/ethClient"
	"base/logger"
	"base/modules"
	"math/rand"
	"sync"
	"time"

	"base/app/helpers"
	"base/app/process"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	helpers.PrintStartupMessages()

	accConfigPath, configPath, statePath, err := helpers.AllPathInit()
	if err != nil {
		logger.GlobalLogger.Fatalf("Ошибка инициализации путей конфигурационных файлов: %v", err)
	}

	accounts, accConfig, err := helpers.AccsInit(accConfigPath)
	if err != nil {
		logger.GlobalLogger.Fatalf("ошибка создания аккаунтов: %v", err)
	}

	config, err := cfg.LoadConfig(configPath)
	if err != nil {
		logger.GlobalLogger.Fatalf("ошибка загрузки основного конфига, проверьте его целостность: %v", err)
	}
	logger.GlobalLogger.Info("Основная конфигурация успешно загружена.")

	clients, err := helpers.ClientsInit()
	if err != nil {
		logger.GlobalLogger.Fatal(err)
	}
	defer ethClient.CloseAllClients(clients)

	mods, err := modules.InitializeModules(*config, clients)
	if err != nil {
		logger.GlobalLogger.Fatalf("ошибка инициализации модулей: %v", err)
	}
	logger.GlobalLogger.Info("Все модули успешно инициализированы. Спим 2 секунды.")
	time.Sleep(time.Second * 2)

	memoryHandler := process.NewMemory(statePath)

	process.UploadOldAction(memoryHandler)

	availableNFTs := account.InitializeAvailableNFTs(accConfig)
	randomizer := randomization.NewRandomizer(cfg.AviableTokens, availableNFTs, clients)

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
	logger.GlobalLogger.Info(cfg.Subscribe)
}
