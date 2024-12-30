BINARY_NAME=base

SRC_DIR=./app

ifeq ($(OS),Windows_NT)
    RM=del /F /Q
    BINARY_EXT=.exe
else
    RM=rm -f
    BINARY_EXT=
endif

all: build

build:
	@echo "==> Компиляция проекта..."
	go build -o $(BINARY_NAME)$(BINARY_EXT) $(SRC_DIR)/main.go

run: build
	@echo "==> Запуск приложения..."
	./$(BINARY_NAME)$(BINARY_EXT)

clean:
	@echo "==> Очистка..."
	$(RM) $(BINARY_NAME)$(BINARY_EXT)

deps:
	@echo "==> Установка зависимостей..."
	go mod download

update-deps:
	@echo "==> Обновление зависимостей..."
	go get -u ./...
	go mod tidy

help:
	@echo "Использование: make [команда]"
	@echo ""
	@echo "Доступные команды:"
	@echo "  build          Компилирует проект"
	@echo "  run            Компилирует и запускает проект"
	@echo "  clean          Удаляет скомпилированные файлы"
	@echo "  deps           Устанавливает зависимости"
	@echo "  update-deps    Обновляет зависимости"
	@echo "  help           Показывает эту справку"

.PHONY: all build run clean deps update-deps help