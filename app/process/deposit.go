package process

import (
	"base/account"
	"base/actions/handlers"
	"base/actions/types"
	"base/config"
	"base/ethClient"
	"base/logger"
	"base/modules"
	"base/utils"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

func bridgeToBase(acc *account.Account, mainConfig *config.Config, clients map[string]*ethClient.Client, mods *modules.Modules) error {
	if err := ensureRefuelIfNeeded(acc, clients, mods); err != nil {
		return err
	}

	tokenAddress, err := getBridgeTokenAddress(acc)
	if err != nil {
		return fmt.Errorf("ошибка получения токена для бриджа %v", err)
	}

	amountToBridge, err := calculateBridgeAmount(acc, clients, tokenAddress)
	if err != nil {
		return fmt.Errorf("ошибка расчета бриджа %v", err)
	}

	if err = approveIfNeeded(acc, clients, tokenAddress, amountToBridge); err != nil {
		logger.GlobalLogger.Errorf("ошибка approve: %v", err)
	}

	time.Sleep(time.Second * 5)

	if err := executeBridge(acc, mainConfig, clients, mods, amountToBridge); err != nil {
		logger.GlobalLogger.Warnf("bridge error: %v", err)
	}

	waitAfterBridge(acc)

	return nil
}

func checkRefuel(acc *account.Account, clients map[string]*ethClient.Client, mods *modules.Modules) error {
	needsRefuel, maxChain, err := depositNativeIfNeeded(acc, clients)
	if err != nil {
		logger.GlobalLogger.Errorf("Ошибка проверки необходимости рефьюела: %v", err)
		return err
	}

	if needsRefuel {
		if err := mods.Refuel.Refuel(maxChain, "base", acc); err != nil {
			logger.GlobalLogger.Warnf("Ошибка депозита нативки в base: %v", err)
			return err
		}
	}
	return nil
}

func ensureRefuelIfNeeded(acc *account.Account, clients map[string]*ethClient.Client, mods *modules.Modules) error {
	needsRefuel, maxChain, err := depositNativeIfNeeded(acc, clients)
	if err != nil {
		logger.GlobalLogger.Errorf("Ошибка проверки необходимости рефьюела: %v", err)
		return err
	}
	if needsRefuel {
		if err := mods.Refuel.Refuel(maxChain, "base", acc); err != nil {
			logger.GlobalLogger.Warnf("Ошибка депозита нативки в base: %v", err)
			return err
		}
	}
	return nil
}

func calculateBridgeAmount(acc *account.Account, clients map[string]*ethClient.Client, tokenAddress common.Address) (*big.Int, error) {
	amount, err := clients[acc.Bridge].BalanceCheck(acc.Address, tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed bridge: %v", err)
	}
	reduction := new(big.Int).Div(new(big.Int).Mul(amount, big.NewInt(5)), big.NewInt(100))
	amount.Sub(amount, reduction)
	return amount, nil
}

func approveIfNeeded(acc *account.Account, clients map[string]*ethClient.Client, tokenAddress common.Address, amount *big.Int) error {
	if utils.IsNativeTokenBySymbol(acc.TokenBridge) {
		return nil
	}
	_, err := clients[acc.Bridge].ApproveTx(tokenAddress, config.LZ_Main_CA[acc.Bridge]["swap_ca"], acc.Address, acc.PrivateKey, amount, false)
	return err
}

func executeBridge(acc *account.Account, mainConfig *config.Config, clients map[string]*ethClient.Client, mods *modules.Modules, amountToBridge *big.Int) error {
	return handlers.BridgeHandler.Execute(handlers.BridgeHandler{
		BridgeParams: types.BridgeParams{
			FromChain:      acc.Bridge,
			DstChain:       "base",
			AmountToBridge: amountToBridge,
		},
	}, acc, *mods, clients[acc.Bridge], mainConfig)
}

func waitAfterBridge(acc *account.Account) {
	if acc.Bridge == "polygon" {
		logger.GlobalLogger.Info("бридж окончен, спим 25 минут...")
		time.Sleep(time.Minute * 25)
	} else {
		logger.GlobalLogger.Info("бридж окончен, спим 3 минуты...")
		time.Sleep(time.Minute * 3)
	}
}

func depositNativeIfNeeded(acc *account.Account, clients map[string]*ethClient.Client) (bool, string, error) {
	balanceInBase, err := clients["base"].BalanceCheck(acc.Address, config.WETH)
	if err != nil {
		return false, "", err
	}

	if balanceInBase.Cmp(config.MinBalance) >= 0 {
		logger.GlobalLogger.Infof("бридж нативного токена не нужен. Баланс есть")
		return false, "", nil
	}

	maxChain, maxBal, err := getMaxNativeBalanceChain(acc, clients)
	if err != nil {
		return false, "", err
	}

	if maxBal.Cmp(big.NewInt(0)) == 0 {
		return false, "", errors.New("нет баланса в других сетях")
	}

	return true, maxChain, nil
}

func getMaxNativeBalanceChain(acc *account.Account, clients map[string]*ethClient.Client) (string, *big.Int, error) {
	maxChain := ""
	maxBalance := big.NewInt(0)
	for chain, client := range clients {
		if chain == "base" {
			continue
		}

		bal, err := client.BalanceCheck(acc.Address, config.WETH)
		if err != nil {
			logger.GlobalLogger.Warnf("Ошибка получения баланса для сети %s: %v", chain, err)
			continue
		}

		if bal.Cmp(maxBalance) > 0 {
			maxBalance.Set(bal)
			maxChain = chain
		}
	}

	time.Sleep(time.Second * 5)
	return maxChain, maxBalance, nil
}

func getBridgeTokenAddress(acc *account.Account) (common.Address, error) {
	tokenAddress, exists := config.OtherTokens[fmt.Sprintf("%s_%s", acc.Bridge, acc.TokenBridge)]
	if !exists {
		return common.Address{}, fmt.Errorf("токен '%s' не найден в конфигурации для аккаунта %d", acc.TokenBridge, acc.AccountID)
	}
	return tokenAddress, nil
}
