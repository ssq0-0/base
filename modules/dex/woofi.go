package dex

import (
	"base/account"
	"base/config"
	"base/ethClient"
	"base/utils"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type WooFi struct {
	ABI    *abi.ABI
	Client *ethClient.Client
	CA     common.Address
}

func NewWooFi(client *ethClient.Client, ca, abiPath string) (*WooFi, error) {
	abi, err := utils.ReadAbi(abiPath)
	if err != nil {
		return nil, err
	}

	return &WooFi{
		ABI:    abi,
		CA:     common.HexToAddress(ca),
		Client: client,
	}, nil
}

func (wf *WooFi) Swap(fromToken, toToken common.Address, amountIn, value *big.Int, acc *account.Account) error {
	amountMinOut, err := wf.querySwap(fromToken, toToken, amountIn)
	if err != nil {
		return err
	}

	data, err := wf.ABI.Pack("swap", fromToken, toToken, amountIn, amountMinOut, acc.Address, acc.Address)
	if err != nil {
		return err
	}

	return wf.Client.SendTransaction(acc.PrivateKey, acc.Address, wf.CA, wf.Client.GetNonce(acc.Address), value, data)
}

func (wf *WooFi) querySwap(fromToken, toToken common.Address, amountIn *big.Int) (*big.Int, error) {
	data, err := wf.ABI.Pack("tryQuerySwap", fromToken, toToken, amountIn)
	if err != nil {
		return nil, err
	}

	return getAmountMin(wf.CA, data, wf.Client, wf.ABI, "tryQuerySwap", config.Slippage)
}
