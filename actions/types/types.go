package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type ActionType string

type LiquidParams struct {
	Type   string
	Token  common.Address
	Amount *big.Int
}

type BSNParams struct {
	Name string
}

type DexParams struct {
	FromToken    common.Address
	ToToken      common.Address
	AmountToSwap *big.Int
}
type BridgeParams struct {
	FromChain      string
	DstChain       uint16
	SrcPoolId      *big.Int // Используется для Stargate
	DstPoolId      *big.Int // Используется для Stargate
	AmountToBridge *big.Int
}

type NftMintParams struct {
	MintCA common.Address
	Price  *big.Int
}

const (
	BridgeAction           ActionType = "stargate"
	UniswapAction          ActionType = "uniswap"
	PancakeAction          ActionType = "pancake"
	WoofiAction            ActionType = "woofi"
	ZoraAction             ActionType = "zora"
	NFT2MeAction           ActionType = "nft2me"
	BaseNameAction         ActionType = "basenames"
	DmailAction            ActionType = "dmail"
	AaveETHDepositAction   ActionType = "aave_deposit"
	AaveETHWithdrawAction  ActionType = "aave_withdraw"
	AaveUSDCSupplyAction   ActionType = "aave_supply"
	AaveUSDCWithdrawAction ActionType = "aave_withdraw_usdc"
	MoonwellDepositAction  ActionType = "moonwell_deposit"
	MoonwellWithdrawAction ActionType = "moonwell_withdraw"
	CollectorModAction     ActionType = "collector_mod"
)
