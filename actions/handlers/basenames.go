package handlers

import (
	"base/account"
	"base/config"
	"base/ethClient"
	"base/modules"
	"context"
	"errors"
	"math/big"
)

type BaseNameHandler struct {
}

func (bh BaseNameHandler) Execute(acc *account.Account, mods modules.Modules, client *ethClient.Client, config *config.Config) error {
	price, err := bh.calculatePrice(acc.BaseName)
	if err != nil {
		return err
	}

	balance, err := client.Client.BalanceAt(context.Background(), acc.Address, nil)
	if err != nil {
		return err
	}

	if balance.Cmp(price) < 0 {
		return errors.New("insufficient balance to register a name")
	}

	return mods.Domains.RegisterName(acc.BaseName, price, acc)
}

func (bh BaseNameHandler) calculatePrice(name string) (*big.Int, error) {
	length := len(name)

	var price *big.Float
	switch {
	case length == 3:
		price = config.BSN_price_3length
	case length == 4:
		price = config.BSN_price_4length
	case length >= 5 && length <= 9:
		price = config.BSN_price_5length
	case length >= 10:
		price = config.BSN_price_10length
	default:
		return nil, errors.New("the length of the name must be more than 3 characters")
	}

	weiMultiplier := new(big.Float).SetFloat64(1e18)
	priceWei := new(big.Float).Mul(price, weiMultiplier)

	priceInt := new(big.Int)
	priceWei.Int(priceInt)

	return priceInt, nil
}
