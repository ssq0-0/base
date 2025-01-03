package account

import (
	"encoding/json"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

type RandomConfig struct {
	Wallets      []WalletConfig `json:"wallets"`
	Modules      ModulesConfig  `json:"modules"`
	NFTContracts NFTCategories  `json:"nft_ca"`
}

type WalletConfig struct {
	PrivateKey      string `json:"private_key"`
	Endpoint        string `json:"endpoint"`
	RevertAllowance bool   `json:"revert_allowance"`
	BaseName        string `json:"base_name"`
	NameUsed        bool
	UsedRange       int64  `json:"used_range"`
	PoolUsedRange   int64  `json:"used_range_in_pools"`
	Bridge          string `json:"bridge"`
	Token           string `json:"token"`
	ActionNumMIN    *int   `json:"action_num_min"`
	ActionNumMAX    *int   `json:"action_num_max"`
	ActionTimeMIN   *int   `json:"action_time_window_MIN"`
	ActionTimeMAX   *int   `json:"action_time_window_MAX"`
}

type NFTCategories struct {
	Nft2Me map[string]string `json:"nf2me"`
	Zora   map[string]string `json:"zora"`
}

type ModulesConfig struct {
	Uniswap   bool `json:"uniswap"`
	Pancake   bool `json:"pancake"`
	Woofi     bool `json:"woofi"`
	OpenOcean bool `json:"openocean"`
	Odos      bool `json:"odos"`
	Refuel    bool `json:"refuel"`
	Zora      bool `json:"zora"`
	NFT2Me    bool `json:"nft2me"`
	BaseNames bool `json:"basenames"`
	Stargate  bool `json:"stargate"`
	Dmail     bool `json:"dmail"`
	Aave      bool `json:"aave"`
	Moonwell  bool `json:"moonwell"`
	Collector bool `json:"collector_mod"`
}

func LoadRandomConfig(path string) (*RandomConfig, error) {
	var cfg *RandomConfig
	file, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	return cfg, err
}

func InitializeAvailableNFTs(accConfig *RandomConfig) map[string]map[common.Address]*big.Int {
	availableNFTs := make(map[string]map[common.Address]*big.Int)

	if accConfig.Modules.NFT2Me {
		processNFTCategory("nft2me", accConfig.NFTContracts.Nft2Me, availableNFTs)
	}

	if accConfig.Modules.Zora {
		processNFTCategory("zora", accConfig.NFTContracts.Zora, availableNFTs)
	}

	return availableNFTs
}

func processNFTCategory(protocolName string, contracts map[string]string, availableNFTs map[string]map[common.Address]*big.Int) {
	availableNFTs[protocolName] = make(map[common.Address]*big.Int)

	isZora := strings.ToLower(protocolName) == "zora"

	for addrStr, priceStr := range contracts {
		priceFloat, ok := new(big.Float).SetString(priceStr)
		if !ok {
			log.Printf("Ошибка преобразования строки в float для %s контракта %s: %s", protocolName, addrStr, priceStr)
			continue
		}

		var priceInWei *big.Int
		if isZora {
			sparks := new(big.Int)
			priceFloat.Int(sparks)
			priceInWei = sparksToEth(sparks)
		} else {
			weiMultiplier := new(big.Float).SetFloat64(1e18)
			weiFloat := new(big.Float).Mul(priceFloat, weiMultiplier)
			priceInWei = new(big.Int)
			weiFloat.Int(priceInWei)
		}

		availableNFTs[protocolName][common.HexToAddress(addrStr)] = priceInWei
	}
}

func sparksToEth(sparks *big.Int) *big.Int {
	ethFloat := new(big.Float).SetInt(sparks)
	ethFloat.Mul(ethFloat, big.NewFloat(1e-6))

	weiFloat := new(big.Float).Mul(ethFloat, big.NewFloat(1e18))

	weiInt, _ := weiFloat.Int(nil)

	return weiInt
}
