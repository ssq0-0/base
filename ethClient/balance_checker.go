package ethClient

import (
	"base/config"
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func (c *Client) BalanceCheck(owner, tokenAddr common.Address) (*big.Int, error) {
	if tokenAddr == (common.Address{}) || tokenAddr == config.WETH {
		balance, err := c.Client.BalanceAt(context.Background(), owner, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get native coin balance: %v", err)
		}
		return balance, nil
	}

	data, err := config.Erc20ABI.Pack("balanceOf", owner)
	if err != nil {
		return nil, fmt.Errorf("failed to pack data: %v", err)
	}
	result, err := c.CallCA(tokenAddr, data)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %v", err)
	}

	var balance *big.Int
	if err = config.Erc20ABI.UnpackIntoInterface(&balance, "balanceOf", result); err != nil {
		return nil, fmt.Errorf("failed to unpack result: %v", err)
	}

	return balance, nil
}

func GetTokenDecimals(tokenAddress common.Address, client *ethclient.Client) (int, error) {
	decimals, ok := config.TokenDecimals[tokenAddress]
	if !ok {
		return 0, errors.New("no suitable token found to get decimals")
	}

	return int(decimals), nil
}

func (c *Client) NormalizeBalance(balance *big.Int, token common.Address) (*big.Float, error) {
	decimals, ok := config.TokenDecimals[token]
	if !ok {
		return nil, fmt.Errorf("decimals not found for token %s", token.Hex())
	}

	price, ok := config.TokenPrice[token]
	if !ok {
		return nil, fmt.Errorf("price not found for token %s", token.Hex())
	}

	normalized := new(big.Float).SetInt(balance)
	divisor := new(big.Float).SetFloat64(math.Pow10(int(decimals)))
	normalized.Quo(normalized, divisor)

	normalized.Mul(normalized, price)

	return normalized, nil
}
