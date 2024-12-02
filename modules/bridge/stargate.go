package bridge

import (
	"base/account"
	"base/ethClient"
	"base/models"
	"base/utils"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type Stargate struct {
	SwapABI   *abi.ABI
	FeeABI    *abi.ABI
	Addresses map[string]map[string]common.Address
	Clients   map[string]*ethClient.Client
	PoolsIds  map[string]int
}

func NewStargate(client map[string]*ethClient.Client, addresses map[string]map[string]common.Address, swapABIPath, feeABIPah string) (*Stargate, error) {
	swapABI, err := utils.ReadAbi(swapABIPath)
	if err != nil {
		return nil, err
	}
	feeABI, err := utils.ReadAbi(feeABIPah)
	if err != nil {
		return nil, err
	}

	return &Stargate{
		SwapABI:   swapABI,
		FeeABI:    feeABI,
		Addresses: addresses,
		Clients:   client,
	}, nil
}

func (stg *Stargate) SwapStable(from string, dstChain uint16, srcPoolId, dstPoolId, amountIn *big.Int, acc *account.Account) error {
	fee, err := getFee(stg.Clients[from], stg.Addresses[from]["fee_ca"], stg.FeeABI, acc.Address, dstChain, "quoteLayerZeroFee")
	if err != nil {
		return err
	}

	minAmountLD := calculateMinAmountLD(amountIn)
	swapData, err := stg.SwapABI.Pack(
		"swap",
		dstChain,
		srcPoolId,
		dstPoolId,
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

	return stg.Clients[from].SendTransaction(acc.PrivateKey, acc.Address, stg.Addresses[from]["swap_ca"], stg.Clients[from].GetNonce(acc.Address), fee, swapData)
}
