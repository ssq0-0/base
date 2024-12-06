package account

import (
	"base/config"
	"base/logger"
	"base/models"
	"base/utils"
	"crypto/ecdsa"
	"fmt"
	"math/rand"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sync/errgroup"
)

type Account struct {
	AccountID        int
	Address          common.Address
	Endpoint         common.Address
	RevertAllowance  bool
	PrivateKey       *ecdsa.PrivateKey
	Bridge           string
	BridgeChain      string
	TokenBridge      string
	BaseName         string
	UsedRange        int64
	PoolUsedRange    int64
	LastSwaps        []models.SwapPair
	LastBridge       string
	LastTokenDeposit common.Address
	LastToken        common.Address
	LastMint         []common.Address
	LastPoolAction   []string
	LastPool         string
	ActionNumMin     int
	ActionNumMax     int
	ActionTimeMIN    int
	ActionTimeMAX    int
}

func NewAccount(accountID int, privateKey *ecdsa.PrivateKey, address common.Address, endpoint, baseName string, revert bool, usedRange, poolUsedRange int64, bridge, tokenBridge string, actionNumMin, actionNumMax, actionTimeMIN, actionTimeMAX int) *Account {
	return &Account{
		AccountID:       accountID,
		PrivateKey:      privateKey,
		RevertAllowance: revert,
		Bridge:          bridge,
		TokenBridge:     tokenBridge,
		BaseName:        baseName,
		Address:         address,
		Endpoint:        common.HexToAddress(endpoint),
		UsedRange:       usedRange,
		PoolUsedRange:   poolUsedRange,
		LastMint:        []common.Address{},
		LastPoolAction:  []string{},
		ActionNumMin:    actionNumMin,
		ActionNumMax:    actionNumMax,
		ActionTimeMIN:   actionTimeMIN,
		ActionTimeMAX:   actionTimeMAX,
	}
}

func CreateAccounts(walletConfigs []WalletConfig) ([]*Account, error) {
	var (
		accounts     []*Account
		accountsLock sync.Mutex
	)

	g := new(errgroup.Group)

	for idx, wc := range walletConfigs {
		idx := idx
		wc := wc
		g.Go(func() error {
			privateKey, err := utils.ParsePrivateKey(wc.PrivateKey)
			if err != nil {
				return fmt.Errorf("ошибка парсинга приватного ключа для кошелька %d: %v", idx+1, err)
			}

			address := utils.DeriveAddress(privateKey)

			if wc.UsedRange == 0 {
				wc.UsedRange = int64(70 + rand.Intn(31))
			}
			if wc.PoolUsedRange == 0 {
				wc.PoolUsedRange = int64(20 + rand.Intn(31))
			}
			if wc.ActionNumMIN == nil && wc.ActionNumMAX == nil {
				wc.ActionNumMIN = &config.DEFAULT_actionNumMin
				wc.ActionNumMAX = &config.DEFAULT_actionNumMax
			}
			if wc.ActionTimeMIN == nil && wc.ActionTimeMAX == nil {
				wc.ActionTimeMIN = &config.DEFAULT_actionTimeMin
				wc.ActionTimeMAX = &config.DEFAULT_actionTimeMax
			}

			account := NewAccount(
				idx+1,
				privateKey,
				address,
				wc.Endpoint,
				wc.BaseName,
				wc.RevertAllowance,
				wc.UsedRange,
				wc.PoolUsedRange,
				wc.Bridge,
				wc.Token,
				*wc.ActionNumMIN,
				*wc.ActionNumMAX,
				*wc.ActionTimeMIN,
				*wc.ActionTimeMAX,
			)

			accountsLock.Lock()
			accounts = append(accounts, account)
			accountsLock.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		logger.GlobalLogger.Error(err)
	}

	if len(accounts) == 0 {
		return nil, fmt.Errorf("нет созданных аккаунтов. Проверьте приватные ключи в конфигурации")
	}

	return accounts, nil
}
