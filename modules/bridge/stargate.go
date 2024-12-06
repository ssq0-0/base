package bridge

import (
	"base/account"
	"base/ethClient"
	"base/models"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type Stargate struct {
	SwapABI  *abi.ABI
	FeeABI   *abi.ABI
	SwapCAs  map[string]common.Address
	FeeCAs   map[string]common.Address
	Clients  map[string]*ethClient.Client
	PoolsIds map[string]*big.Int
	ChainIDs map[string]uint16
}

func NewStargate(client map[string]*ethClient.Client, swapCAs, feeCAs map[string]common.Address, poolIds map[string]*big.Int, chainIds map[string]uint16, swapABI, feeABI *abi.ABI) (*Stargate, error) {
	return &Stargate{
		SwapABI:  swapABI,
		FeeABI:   feeABI,
		SwapCAs:  swapCAs,
		FeeCAs:   feeCAs,
		Clients:  client,
		PoolsIds: poolIds,
		ChainIDs: chainIds,
	}, nil
}

func (stg *Stargate) SwapStable(from, dstChain, token string, amountIn *big.Int, acc *account.Account) error {
	fee, err := getFee(stg.Clients[from], stg.FeeCAs[from], stg.FeeABI, acc.Address, stg.ChainIDs[dstChain], "quoteLayerZeroFee")
	if err != nil {
		return err
	}

	minAmountLD := calculateMinAmountLD(amountIn)
	swapData, err := stg.SwapABI.Pack(
		"swap",
		stg.ChainIDs[dstChain],
		stg.PoolsIds[fmt.Sprintf("%s_%s", from, token)],
		stg.PoolsIds["base_usdc"],
		acc.Address,
		amountIn,
		minAmountLD,
		models.LzTxObj{
			DstGasForCall:   big.NewInt(0),
			DstNativeAmount: big.NewInt(0),
			DstNativeAddr:   common.Hex2Bytes("0000000000000000000000000000000000000001"),
		},
		acc.Address.Bytes(),
		[]byte{},
	)
	if err != nil {
		return fmt.Errorf("failed pack data for stargate: %v", err)
	}

	return stg.Clients[from].SendTransaction(acc.PrivateKey, acc.Address, stg.SwapCAs[from], stg.Clients[from].GetNonce(acc.Address), fee, swapData)
}
