package ethClient

import (
	"base/config"
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func (c *Client) ApproveTx(tokenAddr, spender, ownerAddr common.Address, privateKey *ecdsa.PrivateKey, amount *big.Int, rollback bool) (*types.Transaction, error) {
	if isNativeToken(tokenAddr) {
		return nil, nil
	}

	allowance, err := c.Allowance(tokenAddr, ownerAddr, spender)
	if err != nil {
		return nil, fmt.Errorf("failed to get allowance: %v", err)
	}

	var approveValue *big.Int
	if rollback {
		approveValue = big.NewInt(0)
	} else {
		if allowance.Cmp(amount) >= 0 {
			log.Printf("The current allowance %s is sufficient for the amount of %s, sending an approve transaction is not required.", allowance.String(), amount.String())
			return nil, nil
		}
		approveValue = config.MaxUint256
	}

	approveData, err := config.Erc20ABI.Pack("approve", spender, approveValue)
	if err != nil {
		return nil, fmt.Errorf("failed to pack approve data: %v", err)
	}

	gasLimit, maxPriorityFeePerGas, maxFeePerGas, err := c.GetGasValues(ethereum.CallMsg{
		From: ownerAddr,
		To:   &tokenAddr,
		Data: approveData,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %v", err)
	}

	chainID, err := c.Client.NetworkID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get ChainID: %v", err)
	}

	nonce, err := c.Client.PendingNonceAt(context.Background(), ownerAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %v", err)
	}

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: maxPriorityFeePerGas,
		GasFeeCap: maxFeePerGas,
		Gas:       gasLimit,
		To:        &tokenAddr,
		Value:     big.NewInt(0),
		Data:      approveData,
	})

	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %v", err)
	}

	if err := c.Client.SendTransaction(context.Background(), signedTx); err != nil {
		return nil, fmt.Errorf("failed to send transaction: %v", err)
	}

	log.Printf("Approve transaction sent: https://explorer.chainid-%d.org/tx/%s", chainID, signedTx.Hash().Hex())
	time.Sleep(time.Second * 12)
	return signedTx, nil
}

func (c *Client) Allowance(tokenAddr, owner, spender common.Address) (*big.Int, error) {
	data, err := config.Erc20ABI.Pack("allowance", owner, spender)
	if err != nil {
		return nil, fmt.Errorf("failed to pack allowance data: %v", err)
	}

	msg := ethereum.CallMsg{
		To:   &tokenAddr,
		Data: data,
	}

	result, err := c.Client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %v", err)
	}

	var allowance *big.Int
	if err = config.Erc20ABI.UnpackIntoInterface(&allowance, "allowance", result); err != nil {
		return nil, fmt.Errorf("failed to unpack allowance data: %v", err)
	}

	return allowance, nil
}

func isNativeToken(tokenAddr common.Address) bool {
	if tokenAddr == (common.Address{}) {
		return true
	}

	nativeTokens := map[common.Address]bool{
		config.WETH:     true,
		config.WooFiETH: true,
	}

	return nativeTokens[tokenAddr]
}
