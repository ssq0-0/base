package moonwell

import (
	"base/account"
	"base/config"
	"base/ethClient"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type Moonwell struct {
	WethRouter    common.Address
	MoonwellEthCA common.Address
	CA            common.Address
	MWethAbi      *abi.ABI
	ABI           *abi.ABI
	Client        *ethClient.Client
}

func NewMoonwell(client *ethClient.Client, wethRouter, methca common.Address, abi, mwETHAbi *abi.ABI) (*Moonwell, error) {
	return &Moonwell{
		WethRouter:    wethRouter,
		MoonwellEthCA: methca,
		ABI:           abi,
		MWethAbi:      mwETHAbi,
		Client:        client,
	}, nil
}

func (m *Moonwell) DepositETH(amountIn *big.Int, acc *account.Account) error {
	data, err := m.ABI.Pack("mint", acc.Address)
	if err != nil {
		return err
	}

	return m.Client.SendTransaction(acc.PrivateKey, acc.Address, m.WethRouter, m.Client.GetNonce(acc.Address), amountIn, data)
}

func (m *Moonwell) WithdrawETH(acc *account.Account, tokenOut common.Address) error {
	balanceForWithdraw, err := m.Client.BalanceCheck(acc.Address, config.MoonwellWETH)
	if err != nil {
		return nil
	}
	if balanceForWithdraw == nil {
		return fmt.Errorf("обернутый токен Moonwell отсутствует")
	}

	data, err := m.MWethAbi.Pack("redeem", balanceForWithdraw)
	if err != nil {
		return err
	}

	return m.Client.SendTransaction(acc.PrivateKey, acc.Address, m.MoonwellEthCA, m.Client.GetNonce(acc.Address), big.NewInt(0), data)
}
