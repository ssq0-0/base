package handlers

import (
	"base/account"
	"base/ethClient"
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

func CalculatePercentageOfBalance(acc *account.Account, client *ethClient.Client, token common.Address, percentage int64, ethAddresses []common.Address) (*big.Int, error) {
	balance, err := getTokenBalance(acc, client, token, ethAddresses)
	if err != nil {
		return nil, err
	}

	amount := new(big.Int).Mul(balance, big.NewInt(percentage))
	amount.Div(amount, big.NewInt(100))

	return amount, nil
}

func getTokenBalance(acc *account.Account, client *ethClient.Client, token common.Address, ethAddresses []common.Address) (*big.Int, error) {
	isETH := false
	for _, ethAddr := range ethAddresses {
		if token == ethAddr {
			isETH = true
			break
		}
	}

	if isETH {
		balanceWei, err := client.Client.BalanceAt(context.Background(), acc.Address, nil)
		if err != nil {
			return nil, fmt.Errorf("failed get native balance: %v", err)
		}
		return balanceWei, nil
	} else {
		erc20Balance, err := client.BalanceCheck(acc.Address, token)
		if err != nil {
			return nil, fmt.Errorf("failed get balance erc20 token: %v", err)
		}
		return erc20Balance, nil
	}
}
