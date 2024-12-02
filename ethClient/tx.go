package ethClient

import (
	"base/logger"
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

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

	log.Printf("Transaction sent: https://explorer.chainid-%d.org/tx/%s", chainID, signedTx.Hash().Hex())

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

	signer := types.LatestSignerForChainID(chainID)

	from, err := types.Sender(signer, tx)
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
