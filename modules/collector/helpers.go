package collector

import (
	"base/account"
	"base/logger"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

func (c *Collector) checkAndNormalizeBalance(acc *account.Account, token common.Address, mu *sync.Mutex) (*big.Int, bool) {
	balance, err := c.Client.BalanceCheck(acc.Address, token)
	if err != nil {
		logger.GlobalLogger.Errorf("Ошибка получения баланса для токена %s", token.Hex())
		return nil, false
	}

	mu.Lock()
	normalizedBalance, err := c.Client.NormalizeBalance(balance, token)
	mu.Unlock()
	if err != nil {
		logger.GlobalLogger.Errorf("Ошибка получения цены для токена %s", token.Hex())
		return nil, false
	}

	if normalizedBalance.Cmp(c.minBalanceUSD) <= 0 {
		logger.GlobalLogger.Infof("Пропускаем токен %s с балансом $%.2f (меньше $%.2f)", token.Hex(), normalizedBalance, c.minBalanceUSD)
		return balance, false
	}

	logger.GlobalLogger.Infof("Обрабатываем токен %s с балансом $%.2f", token.Hex(), normalizedBalance)
	return balance, true
}
