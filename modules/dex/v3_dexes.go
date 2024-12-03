package dex

import (
	"base/account"
	"base/config"
	"base/ethClient"
	"base/models"
	"base/utils"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type V3Router struct {
	RouterABI         *abi.ABI
	QuoterABI         *abi.ABI
	Client            *ethClient.Client
	RouterCA          common.Address
	QuoterCA          common.Address
	Fee               *big.Int
	SqrtPriceLimitX96 *big.Int
}

func NewV3Router(client *ethClient.Client, RouterCA, QuoterCA, RouterABIPath, QuoterABIPath string, fee, sqrtPriceLimitX96 *big.Int) (*V3Router, error) {
	routerABI, err := utils.ReadAbi(RouterABIPath)
	if err != nil {
		return nil, err
	}
	quoterABI, err := utils.ReadAbi(QuoterABIPath)
	if err != nil {
		return nil, err
	}

	return &V3Router{
		RouterABI:         routerABI,
		QuoterABI:         quoterABI,
		RouterCA:          common.HexToAddress(RouterCA),
		QuoterCA:          common.HexToAddress(QuoterCA),
		Client:            client,
		Fee:               fee,
		SqrtPriceLimitX96: sqrtPriceLimitX96,
	}, nil
}

func (v3 *V3Router) Swap(fromToken, toToken common.Address, amountIn, value *big.Int, acc *account.Account) error {
	data, _, err := v3.prepareSwapData(acc.Address, fromToken, toToken, amountIn)
	if err != nil {
		return err
	}

	return v3.Client.SendTransaction(acc.PrivateKey, acc.Address, v3.RouterCA, v3.Client.GetNonce(acc.Address), value, data)
}

func (v3 *V3Router) SwapToETH(fromToken, toToken common.Address, amountIn, value *big.Int, acc *account.Account) error {
	data, amountMinOut, err := v3.prepareSwapData(v3.RouterCA, fromToken, toToken, amountIn)
	if err != nil {
		return err
	}

	unwrapData, err := v3.RouterABI.Pack("unwrapWETH9", amountMinOut, acc.Address)
	if err != nil {
		return fmt.Errorf("data packing error for unwrapWETH9: %w", err)
	}

	txData, err := v3.RouterABI.Pack("multicall", [][]byte{data, unwrapData})
	if err != nil {
		return fmt.Errorf("data packing error for multicall: %w", err)
	}

	return v3.Client.SendTransaction(acc.PrivateKey, acc.Address, v3.RouterCA, v3.Client.GetNonce(acc.Address), value, txData)
}

func (v3 *V3Router) prepareSwapData(recipient, fromToken, toToken common.Address, amountIn *big.Int) ([]byte, *big.Int, error) {
	amountMinOut, err := v3.getQuoteSingle(fromToken, toToken, v3.Fee, amountIn)
	if err != nil {
		return nil, nil, fmt.Errorf("error of receiving a quote: %w", err)
	}

	if amountMinOut.Cmp(big.NewInt(0)) <= 0 {
		return nil, nil, fmt.Errorf("minimum output amount is zero")
	}

	data, err := v3.packTxData(recipient, fromToken, toToken, v3.Fee, amountIn, amountMinOut, v3.SqrtPriceLimitX96, v3.RouterABI)
	if err != nil {
		return nil, nil, fmt.Errorf("data packaging error for swap:: %w", err)
	}

	return data, amountMinOut, nil
}

func (v3 *V3Router) getQuoteSingle(fromToken, toToken common.Address, feeOrTickSpacing, amountIn *big.Int) (*big.Int, error) {
	data, err := v3.packQuoteData(fromToken, toToken, feeOrTickSpacing, amountIn, v3.QuoterABI)
	if err != nil {
		return nil, fmt.Errorf("failed pack data for quoteExactInputSingle: %w", err)
	}

	return getAmountMin(v3.QuoterCA, data, v3.Client, v3.QuoterABI, "quoteExactInputSingle", config.Slippage)
}

func (v3 *V3Router) packTxData(ownerAddr, fromToken, toToken common.Address, feeOrTickSpacing, amountIn, amountMinOut, sqrtPriceLimitX96 *big.Int, routerABI *abi.ABI) ([]byte, error) {
	return routerABI.Pack("exactInputSingle", models.ExactInputSingleParams{
		TokenIn:           fromToken,
		TokenOut:          toToken,
		Fee:               feeOrTickSpacing,
		Recipient:         ownerAddr,
		AmountIn:          amountIn,
		AmountOutMinimum:  amountMinOut,
		SqrtPriceLimitX96: sqrtPriceLimitX96,
	})
}

func (v3 *V3Router) packQuoteData(fromToken, toToken common.Address, feeOrTickSpacing, amountIn *big.Int, quoterABI *abi.ABI) ([]byte, error) {
	return quoterABI.Pack("quoteExactInputSingle", models.OtherDexParams{
		TokenIn:           fromToken,
		TokenOut:          toToken,
		AmountIn:          amountIn,
		Fee:               feeOrTickSpacing,
		SqrtPriceLimitX96: big.NewInt(0),
	})
}
