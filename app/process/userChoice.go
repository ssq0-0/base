package process

import (
	"base/logger"
	"fmt"
)

func UploadOldAction(memoryHandler *Memory) {
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
			err := memoryHandler.ClearAllStates()
			if err != nil {
				logger.GlobalLogger.Errorf("Ошибка очистки состояния: %v", err)
				return
			}
		}
	} else {
		logger.GlobalLogger.Info("Файл состояния пуст. Начинаем выполнение с чистого листа.")
	}
}
