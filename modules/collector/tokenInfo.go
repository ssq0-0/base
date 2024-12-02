package collector

import "github.com/ethereum/go-ethereum/common"

type TokenType int

const (
	ERC20 TokenType = iota
	AaveLiquidityPool
	MoonwellLiquidityPool
)

type TokenInfo struct {
	Address       common.Address
	Type          TokenType
	Pool          interface{}
	RequiresPrice bool
}
