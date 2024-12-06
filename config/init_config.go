package config

import (
	"base/logger"
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

type Config struct {
	DexConfig         DexConfig         `json:"dex"`
	BridgeConfig      BridgeConfig      `json:"bridge"`
	RefuelConfig      RefuelConfig      `json:"refuel"`
	DomainsConfig     DomainsConfig     `json:"domains"`
	DmailConfig       DmailConfig       `json:"dmail"`
	LiquidPoolsConfig LiquidPoolsConfig `json:"liquid_pools"`
	NFTMintsConfig    NFTMintsConfig    `json:"nft_mints"`
}

type DexConfig struct {
	Uniswap           V3RouterConfig `json:"uniswap"`
	Pancake           V3RouterConfig `json:"pancake"`
	Woofi             WoofiConfig    `json:"woofi"`
	SqrtPriceLimitX96 *big.Int       `json:"sqrtPriceLimitX96"` // default - 0
}

type V3RouterConfig struct {
	RouterCA      string   `json:"router_ca"`
	QuoterCA      string   `json:"quoter_ca"`
	RouterABIPath string   `json:"router_abi_path"`
	QuoterABIPath string   `json:"quoter_abi_path"`
	Fee           *big.Int `json:"fee"` // default - 0.05%
}

type WoofiConfig struct {
	CA      string `json:"ca"`
	ABIPath string `json:"abi_path"`
}

type BridgeConfig struct {
	SwapAddresses map[string]string `json:"swap_ca"`
	FeeAdresses   map[string]string `json:"fee_ca"`
	SwapABIPath   string            `json:"swap_abi_path"`
	FeeABIPath    string            `json:"fee_abi_path"`
}

type RefuelConfig struct {
	ABIPath         string `json:"api_path"`
	OptimismSocket  string `json:"optimism_socket"`
	ArbitrumSocket  string `json:"arbirum_socket"`
	AvalancheSocket string `json:"avalanche_socket"`
	PolygonSocket   string `json:"polygon_socket"`
	BaseSocket      string `json:"base_socket"`
}

type DomainsConfig struct {
	RegisterCA      string `json:"register_ca"`
	ResolverCA      string `json:"resolver_ca"`
	RegisterABIPath string `json:"register_abi_path"`
	ResolverABIPath string `json:"resolver_abi_path"`
}

type DmailConfig struct {
	CA      string `json:"ca"`
	ABIPath string `json:"abi_path"`
}

type LiquidPoolsConfig struct {
	Aave     AaveConfig     `json:"aave"`
	Moonwell MoonwellConfig `json:"moonwell"`
}

type AaveConfig struct {
	ProxyBase string `json:"proxy_base"`
	EthPool   string `json:"eth_pool"`
	ABIPath   string `json:"abi_path"`
}

type MoonwellConfig struct {
	CA           string `json:"weth_router"`
	METHCA       string `json:"meth_ca"`
	ABIPath      string `json:"abi_path"`
	MWethABIPath string `json:"mweth_abi_path"`
}

type NFTMintsConfig struct {
	Zora   NFTConfig `json:"zora"`
	NFT2Me NFTConfig `json:"nft2me"`
}

type NFTConfig struct {
	CA      string `json:"ca"`
	ABIPath string `json:"abi_path"`
}

func init() {
	MaxUint256, _ = new(big.Int).SetString(MaxUint256Str, 10)

	parsedABI, err := abi.JSON(bytes.NewReader(Erc20JSON))
	if err != nil {
		logger.GlobalLogger.Fatalf("Ошибка при парсинге ABI: %v", err)
	}

	Erc20ABI = &parsedABI

}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	return &cfg, nil
}
