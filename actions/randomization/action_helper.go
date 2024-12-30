package randomization

import (
	"base/account"
	"base/actions/types"
	"crypto/rand"
	"errors"
	"math/big"
)

func getNumActions(walletCfg *account.WalletConfig) (int, error) {
	min := 15
	if walletCfg.ActionNumMIN != nil {
		min = *walletCfg.ActionNumMIN
	}

	max := 25
	if walletCfg.ActionNumMAX != nil {
		max = *walletCfg.ActionNumMAX
	}

	if min > max {
		return 0, errors.New("ActionNumMIN не может быть больше ActionNumMAX")
	}

	return generateRandomInt(min, max)
}

func generateRandomInt(min, max int) (int, error) {
	diff := max - min + 1
	randomBig, err := rand.Int(rand.Reader, big.NewInt(int64(diff)))
	if err != nil {
		return 0, err
	}
	return int(randomBig.Int64()) + min, nil
}

func updateActionHistory(lastActions []string, actionType types.ActionType) []string {
	if len(lastActions) >= 1 {
		lastActions = lastActions[1:]
	}
	lastActions = append(lastActions, string(actionType))
	return lastActions
}
