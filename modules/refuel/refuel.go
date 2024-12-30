package refuel

import (
	"base/account"
	"base/config"
	"base/ethClient"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type Refuel struct {
	Clients   map[string]*ethClient.Client
	Addresses map[string]common.Address
	ChainsIDs map[string]*big.Int
	ABI       *abi.ABI
}

func NewRefuel(clients map[string]*ethClient.Client, adresses map[string]common.Address, chainsIDs map[string]*big.Int, abi *abi.ABI) (*Refuel, error) {
	if len(adresses) == 0 {
		return nil, errors.New("failed init refuel. Check CAs in config")
	}
	return &Refuel{
		Clients:   clients,
		Addresses: adresses,
		ChainsIDs: chainsIDs,
		ABI:       abi,
	}, nil
}

func (rf *Refuel) Refuel(srcChain, dstChain string, acc *account.Account) error {
	amount, err := rf.CheckAndCalculateAmount(srcChain, dstChain, acc)
	if err != nil {
		return err
	}

	data, err := rf.ABI.Pack("depositNativeToken", rf.ChainsIDs[dstChain], acc.Address)
	if err != nil {
		return errors.New("failed pack data for refuel")
	}

	return rf.Clients[srcChain].SendTransaction(acc.PrivateKey, acc.Address, rf.Addresses[srcChain], rf.Clients[srcChain].GetNonce(acc.Address), amount, data)
}

func (rf *Refuel) CheckAndCalculateAmount(srcChain, dstChain string, acc *account.Account) (*big.Int, error) {
	balance, err := rf.Clients[srcChain].BalanceCheck(acc.Address, config.WETH)
	if err != nil {
		return nil, errors.New("не удалось получить баланс на исходной сети: " + err.Error())
	}

	amountToBridge, err := rf.getBridgeAmount(srcChain)
	if err != nil {
		return nil, errors.New("не удалось рассчитать сумму для бриджа: " + err.Error())
	}

	if balance.Cmp(amountToBridge) < 0 {
		return nil, errors.New("недостаточно средств для выполнения бриджа")
	}

	return amountToBridge, nil
}
