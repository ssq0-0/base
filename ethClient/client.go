package ethClient

import (
	"base/config"
	"base/logger"
	"base/utils"
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Client struct {
	Client *ethclient.Client
}

func NewClient(rpc string) (*Client, error) {
	client, err := ethclient.Dial(rpc)
	if err != nil {
		return nil, err
	}

	return &Client{
		Client: client,
	}, nil
}

func CloseAllClients(clients map[string]*Client) {
	for _, client := range clients {
		if client.Client != nil {
			client.Client.Close()
		}
	}
}

func (c *Client) BalanceCheck(owner, tokenAddr common.Address) (*big.Int, error) {
	if utils.IsNativeToken(tokenAddr) {
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
	if err := config.Erc20ABI.UnpackIntoInterface(&balance, "balanceOf", result); err != nil {
		return nil, fmt.Errorf("failed to unpack result: %v", err)
	}

	return balance, nil
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

func (c *Client) CallCA(toCA common.Address, data []byte) ([]byte, error) {
	callMsg := ethereum.CallMsg{
		To:   &toCA,
		Data: data,
	}

	return c.Client.CallContract(context.Background(), callMsg, nil)
}

func (c *Client) GetGasValues(msg ethereum.CallMsg) (uint64, *big.Int, *big.Int, error) {
	header, err := c.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return 0, nil, nil, err
	}
	baseFee := header.BaseFee

	maxPriorityFeePerGas := big.NewInt(1e7)
	maxFeePerGas := new(big.Int).Add(baseFee, maxPriorityFeePerGas)

	gasLimit, err := c.Client.EstimateGas(context.Background(), msg)
	if err != nil {
		return 0, nil, nil, err
	}

	return gasLimit, maxPriorityFeePerGas, maxFeePerGas, nil
}

func (c *Client) GetNonce(address common.Address) uint64 {
	nonce, err := c.Client.PendingNonceAt(context.Background(), address)
	if err != nil {
		return 0
	}
	return nonce
}

func (c *Client) ApproveTx(tokenAddr, spender, ownerAddr common.Address, privateKey *ecdsa.PrivateKey, amount *big.Int, rollback bool) (*types.Transaction, error) {
	if utils.IsNativeToken(tokenAddr) {
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
			logger.GlobalLogger.Infof("The current allowance %s is sufficient for the amount of %s, sending an approve transaction is not required.", allowance.String(), amount.String())
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

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     c.GetNonce(ownerAddr),
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
	if err := c.waitForTransactionSuccess(signedTx.Hash(), 2*time.Minute); err != nil {
		return nil, fmt.Errorf("error waiting for transaction confirmation: %v", err)
	}

	logger.GlobalLogger.Infof("Approve transaction sent: https://basescan.org/tx/%s", chainID, signedTx.Hash().Hex())
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

func (c *Client) SendTransaction(privateKey *ecdsa.PrivateKey, ownerAddr, CA common.Address, nonce uint64, value *big.Int, txData []byte) error {
	chainID, err := c.Client.NetworkID(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get ChainID: %v", err)
	}

	gasLimit, maxPriorityFeePerGas, maxFeePerGas, err := c.GetGasValues(ethereum.CallMsg{
		From:  ownerAddr,
		To:    &CA,
		Value: value,
		Data:  txData,
	})
	if err != nil {
		return fmt.Errorf("failed to estimate gas: %v", err)
	}

	dynamicTx := types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: maxPriorityFeePerGas,
		GasFeeCap: maxFeePerGas,
		Gas:       gasLimit,
		To:        &CA,
		Value:     value,
		Data:      txData,
	}

	signedTx, err := types.SignTx(types.NewTx(&dynamicTx), types.LatestSignerForChainID(chainID), privateKey)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %v", err)
	}

	if err = c.Client.SendTransaction(context.Background(), signedTx); err != nil {
		return fmt.Errorf("failed to send transaction: %v", err)
	}

	log.Printf("Transaction sent: https://basescan.org/tx/%s", signedTx.Hash().Hex())

	return c.waitForTransactionSuccess(signedTx.Hash(), 1*time.Minute)
}

func (c *Client) waitForTransactionSuccess(txHash common.Hash, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return errors.New("transaction wait timeout")
		case <-ticker.C:
			receipt, err := c.Client.TransactionReceipt(context.Background(), txHash)
			if err != nil {
				if err.Error() == "not found" {
					log.Printf("Transaction %s not yet found in the blockchain, retrying...", txHash.Hex())
					continue
				}
				return fmt.Errorf("error getting transaction receipt: %v", err)
			}

			if receipt.Status == 1 {
				logger.GlobalLogger.Infof("Transaction %s succeeded", txHash.Hex())
				return nil
			} else {
				c.logTransactionError(txHash, receipt)
				return errors.New("transaction failed")
			}
		}
	}
}

func (c *Client) logTransactionError(txHash common.Hash, receipt *types.Receipt) {
	logger.GlobalLogger.Errorf("Transaction failed. txHash: %s", txHash.Hex())

	for _, logEntry := range receipt.Logs {
		log.Printf("Event Log - Address: %s, Data: %x, Topics: %v",
			logEntry.Address.Hex(),
			logEntry.Data,
			logEntry.Topics,
		)
	}

	tx, isPending, err := c.Client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		logger.GlobalLogger.Warnf("Error getting transaction details: %v", err)
		return
	}

	chainID, err := c.Client.NetworkID(context.Background())
	if err != nil {
		logger.GlobalLogger.Warnf("Error getting ChainID: %v", err)
		return
	}

	from, err := types.Sender(types.LatestSignerForChainID(chainID), tx)
	if err != nil {
		logger.GlobalLogger.Warnf("Error getting transaction sender: %v", err)
		return
	}

	callMsg := ethereum.CallMsg{
		From:     from,
		To:       tx.To(),
		Gas:      tx.Gas(),
		GasPrice: tx.GasPrice(),
		Value:    tx.Value(),
		Data:     tx.Data(),
	}

	ctx := context.Background()
	blockNumber := receipt.BlockNumber
	result, callErr := c.Client.CallContract(ctx, callMsg, blockNumber)

	var revertReason string

	if callErr != nil {
		if strings.HasPrefix(callErr.Error(), "execution reverted") {
			revertReason = callErr.Error()
			if len(result) > 0 {
				decodedReason, decodeErr := abi.UnpackRevert(result)
				if decodeErr != nil {
					logger.GlobalLogger.Warnf("Failed to decode revert reason: %v", decodeErr)
				} else {
					revertReason = decodedReason
				}
			}
		} else {
			logger.GlobalLogger.Warnf("Error simulating transaction execution: %v", callErr)
		}
	} else {
		if len(result) > 0 {
			decodedReason, decodeErr := abi.UnpackRevert(result)
			if decodeErr != nil {
				logger.GlobalLogger.Warnf("Failed to decode revert reason: %v", decodeErr)
			} else {
				revertReason = decodedReason
			}
		}
	}

	if revertReason != "" {
		logger.GlobalLogger.Warnf("Transaction revert reason: %s", revertReason)
	} else {
		logger.GlobalLogger.Warnf("Transaction revert reason not found.")
	}

	logger.GlobalLogger.Warnf("Failed transaction details:")
	logger.GlobalLogger.Warnf("  From: %s", from.Hex())
	if tx.To() != nil {
		logger.GlobalLogger.Warnf("  To: %s", tx.To().Hex())
	} else {
		logger.GlobalLogger.Warnf("  To: contract creation (contract transaction)")
	}
	logger.GlobalLogger.Warnf("  Value: %s", tx.Value().String())
	logger.GlobalLogger.Warnf("  Gas Limit: %d", tx.Gas())
	logger.GlobalLogger.Warnf("  Gas Price: %s", tx.GasPrice().String())
	logger.GlobalLogger.Warnf("  Nonce: %d", tx.Nonce())
	logger.GlobalLogger.Warnf("  Data: %x", tx.Data())
	logger.GlobalLogger.Warnf("  Pending: %v", isPending)
}
