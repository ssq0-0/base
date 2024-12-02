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
)

type Account struct {
	AccountID        int
	Address          common.Address
	PrivateKey       *ecdsa.PrivateKey
	Bridge           string
	BridgeChain      string
	TokenBridge      string
	BaseName         string
	UsedRange        int64
	PoolUsedRange    int64
	LastSwaps        []models.SwapPair
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

func NewAccount(accountID int, privateKey *ecdsa.PrivateKey, address common.Address, baseName string, usedRange, poolUsedRange int64, bridge, tokenBridge string, actionNumMin, actionNumMax, actionTimeMIN, actionTimeMAX int) *Account {
	return &Account{
		AccountID:      accountID,
		PrivateKey:     privateKey,
		Bridge:         bridge,
		TokenBridge:    tokenBridge,
		BaseName:       baseName,
		Address:        address,
		UsedRange:      usedRange,
		PoolUsedRange:  poolUsedRange,
		LastMint:       []common.Address{},
		LastPoolAction: []string{},
		ActionNumMin:   actionNumMin,
		ActionNumMax:   actionNumMax,
		ActionTimeMIN:  actionTimeMIN,
		ActionTimeMAX:  actionTimeMAX,
	}
}

func CreateAccounts(walletConfigs []WalletConfig) ([]*Account, error) {
	var (
		accounts  []*Account
		wg        sync.WaitGroup
		accountCh = make(chan *Account)
		errorCh   = make(chan error)
	)

	for idx, wc := range walletConfigs {
		wg.Add(1)
		go func(idx int, wc WalletConfig) {
			defer wg.Done()

			privateKey, err := utils.ParsePrivateKey(wc.PrivateKey)
			if err != nil {
				errorCh <- fmt.Errorf("ошибка парсинга приватного ключа для кошелька %d: %v", idx+1, err)
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
				wc.BaseName,
				wc.UsedRange,
				wc.PoolUsedRange,
				wc.Bridge,
				wc.Token,
				*wc.ActionNumMIN,
				*wc.ActionNumMAX,
				*wc.ActionTimeMIN,
				*wc.ActionTimeMAX,
			)

			accountCh <- account

		}(idx, wc)
	}

	go func() {
		wg.Wait()
		close(accountCh)
		close(errorCh)
	}()

	for {
		select {
		case acc, ok := <-accountCh:
			if !ok {
				accountCh = nil
			} else {
				accounts = append(accounts, acc)
			}
		case err, ok := <-errorCh:
			if !ok {
				errorCh = nil
			} else {
				logger.GlobalLogger.Error(err)
			}
		}
		if accountCh == nil && errorCh == nil {
			break
		}
	}

	return accounts, nil
}
