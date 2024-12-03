package collector

import (
	"base/account"
	"base/config"
	"base/ethClient"
	"base/logger"
	"base/modules/dex"
	"base/modules/liquid_pools/aave"
	"base/modules/liquid_pools/moonwell"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type Collector struct {
	availableTokens []TokenInfo
	Client          *ethClient.Client
	Dex             *dex.V3Router
	Aave            *aave.Aave
	Moonwell        *moonwell.Moonwell
	minBalanceUSD   *big.Float
}

func NewCollector(client *ethClient.Client, dex *dex.V3Router, aave *aave.Aave, moonwell *moonwell.Moonwell) *Collector {
	return &Collector{
		availableTokens: []TokenInfo{
			{Address: config.USDC, Type: ERC20, RequiresPrice: true},
			{Address: config.USDbC, Type: ERC20, RequiresPrice: true},
			{Address: config.AaveUSDC, Type: AaveLiquidityPool, Pool: aave, RequiresPrice: false},
			{Address: config.AaveWETH, Type: AaveLiquidityPool, Pool: aave, RequiresPrice: false},
			{Address: config.MoonwellWETH, Type: MoonwellLiquidityPool, Pool: moonwell, RequiresPrice: false},
		},
		Client:        client,
		Dex:           dex,
		Aave:          aave,
		Moonwell:      moonwell,
		minBalanceUSD: config.MinBalanceInDollars,
	}
}

func (c *Collector) Collect(acc *account.Account) error {
	logger.GlobalLogger.Infof("Начало сбора для аккаунта: %s", acc.Address.Hex())

	var mu sync.Mutex
	for _, tokenInfo := range c.availableTokens {
		token := tokenInfo.Address

		switch tokenInfo.Type {
		case ERC20:
			if err := c.processERC20Token(acc, tokenInfo, &mu); err != nil {
				logger.GlobalLogger.Error(err)
				continue
			}
		case AaveLiquidityPool, MoonwellLiquidityPool:
			if err := c.processLiquidityPoolToken(acc, tokenInfo, &mu); err != nil {
				logger.GlobalLogger.Error(err)
				continue
			}
		default:
			logger.GlobalLogger.Errorf("Неизвестный тип токена: %s", token.Hex())
		}
	}

	if err := c.rollbackAllowances(acc); err != nil {
		logger.GlobalLogger.Errorf("Ошибка при откате allowances: %v", err)
	}

	logger.GlobalLogger.Infof("Сбор и откат allowances завершены успешно для аккаунта: %s", acc.Address.Hex())
	return nil
}

func (c *Collector) processERC20Token(acc *account.Account, t TokenInfo, mu *sync.Mutex) error {
	balance, shouldProcess := c.checkAndNormalizeBalance(acc, t.Address, mu)
	if !shouldProcess {
		return nil
	}

	if err := c.ApproveAndSwap(acc, t.Address, balance); err != nil {
		return err
	}

	logger.GlobalLogger.Infof("Своп ERC20 токена %s выполнен успешно", t.Address.Hex())
	return nil
}

func (c *Collector) processLiquidityPoolToken(acc *account.Account, t TokenInfo, mu *sync.Mutex) error {
	balance, shouldProcess := c.checkAndNormalizeBalance(acc, t.Address, mu)
	if !shouldProcess {
		return nil
	}

	switch t.Type {
	case AaveLiquidityPool:
		aavePool, ok := t.Pool.(*aave.Aave)
		if !ok {
			return fmt.Errorf("incorrect pool type for token %s", t.Address.Hex())
		}
		if err := c.handleAaveToken(acc, aavePool, t.Address, balance); err != nil {
			return err
		}
	case MoonwellLiquidityPool:
		moonwellPool, ok := t.Pool.(*moonwell.Moonwell)
		if !ok {
			return fmt.Errorf("incorrect pool type for token %s", t.Address.Hex())
		}
		if err := c.handleMoonwellToken(acc, moonwellPool, t.Address); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown token type for token %s", t.Address.Hex())
	}

	if err := c.ApproveAndSwap(acc, t.Address, balance); err != nil {
		return err
	}

	logger.GlobalLogger.Infof("Своп токена %s выполнен успешно", t.Address.Hex())
	return nil
}

func (c *Collector) rollbackAllowances(acc *account.Account) error {
	logger.GlobalLogger.Infof("Начало отката allowances для аккаунта: %s", acc.Address.Hex())

	for _, tokenInfo := range c.availableTokens {
		token := tokenInfo.Address

		protocols, exists := config.PROTOCOLS_CAs[token]
		if !exists {
			logger.GlobalLogger.Infof("Нет протоколов для отката разрешений для токена %s, пропускаем", token.Hex())
			continue
		}

		for _, protocolCA := range protocols {
			logger.GlobalLogger.Infof("Устанавливаем allowance на ноль для токена %s и протокола %s (%s)", token.Hex(), protocolCA.Hex(), protocolCA.Hex())

			_, err := c.Client.ApproveTx(token, protocolCA, acc.Address, acc.PrivateKey, big.NewInt(0), true)
			if err != nil {
				continue
			}

			logger.GlobalLogger.Infof("Allowance для токена %s и протокола %s успешно откатан на ноль", token.Hex(), protocolCA.Hex())
		}
	}

	logger.GlobalLogger.Infof("Откат allowances завершен успешно для аккаунта: %s", acc.Address.Hex())
	return nil
}

func (c *Collector) handleAaveToken(acc *account.Account, aavePool *aave.Aave, token common.Address, balance *big.Int) error {
	switch token {
	case config.AaveWETH:
		logger.GlobalLogger.Infof("Выводим WETH из Aave для токена %s", token.Hex())
		return aavePool.WithdrawETH(acc, balance)
	case config.AaveUSDC:
		logger.GlobalLogger.Infof("Выводим USDC из Aave для токена %s", token.Hex())
		return aavePool.Withdraw(acc, config.USDC)
	default:
		return fmt.Errorf("неизвестный токен в Aave: %s", token.Hex())
	}
}

func (c *Collector) handleMoonwellToken(acc *account.Account, moonwellPool *moonwell.Moonwell, token common.Address) error {
	switch token {
	case config.MoonwellWETH:
		logger.GlobalLogger.Infof("Выводим WETH из Moonwell для токена %s", token.Hex())
		return moonwellPool.WithdrawETH(acc, config.WETH)
	default:
		return fmt.Errorf("неизвестный токен в Moonwell: %s", token.Hex())
	}
}

func (c *Collector) ApproveAndSwap(acc *account.Account, token common.Address, balance *big.Int) error {
	if balance.Cmp(big.NewInt(0)) == 0 {
		logger.GlobalLogger.Infof("Баланс токена %s равен нулю, пропускаем своп", token.Hex())
		return nil
	}

	_, err := c.Client.ApproveTx(token, c.Dex.RouterCA, acc.Address, acc.PrivateKey, config.MaxUint256, false)
	if err != nil {
		return fmt.Errorf("ошибка аппрува токена %s: %v", token.Hex(), err)
	}

	logger.GlobalLogger.Infof("Свопаем токен %s в ETH, сумма: %s", token.Hex(), balance.String())
	if err := c.Dex.SwapToETH(token, config.WETH, balance, big.NewInt(0), acc); err != nil {
		return fmt.Errorf("ошибка свопа токена %s в ETH: %v", token.Hex(), err)
	}

	logger.GlobalLogger.Infof("Своп токена %s в ETH выполнен успешно, ждем 5 секунд", token.Hex())
	time.Sleep(time.Second * 5)
	return nil
}
