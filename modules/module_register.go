package modules

import (
	"base/config"
	"base/ethClient"
	"base/modules/bridge"
	"base/modules/collector"
	"base/modules/dex"
	"base/modules/dmail"
	"base/modules/domains"
	"base/modules/liquid_pools/aave"
	"base/modules/liquid_pools/moonwell"
	nftmints "base/modules/nft_mints"
	"base/modules/refuel"
	"base/utils"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sync/errgroup"
)

type Modules struct {
	Dex         *DexModules
	Bridge      *bridge.Stargate
	Refuel      *refuel.Refuel
	Dmail       *dmail.Dmail
	Domains     *domains.BSN
	LiquidPools *LiquidPoolsModules
	NFTMints    *NFTMintsModules
	Collector   *collector.Collector
}

type DexModules struct {
	Pancake *dex.V3Router
	Uniswap *dex.V3Router
	Woofi   *dex.WooFi
}

type LiquidPoolsModules struct {
	Aave     *aave.Aave
	Moonwell *moonwell.Moonwell
}

type NFTMintsModules struct {
	Zora   *nftmints.Zora
	NFT2Me *nftmints.Nft2Me
}

func InitializeModules(cfg config.Config, clients map[string]*ethClient.Client) (*Modules, error) {
	var modules Modules
	var g errgroup.Group

	g.Go(func() error {
		var err error
		modules.Dex, err = initializeDexModules(clients["base"], cfg)
		return err
	})

	g.Go(func() error {
		var err error
		modules.Bridge, err = initializeStargate(clients, cfg)
		return err
	})

	g.Go(func() error {
		var err error
		modules.Refuel, err = initializeRefuel(clients, cfg)
		return err
	})

	g.Go(func() error {
		var err error
		modules.LiquidPools, err = initializeLiquidPoolsModules(clients["base"], cfg)
		return err
	})

	g.Go(func() error {
		var err error
		modules.NFTMints, err = initializeNFTMintsModules(clients["base"], cfg)
		return err
	})

	g.Go(func() error {
		var err error
		modules.Dmail, err = dmail.NewDmail(clients["base"], cfg.DmailConfig.CA, cfg.DmailConfig.ABIPath)
		return err
	})

	g.Go(func() error {
		var err error
		modules.Domains, err = initializeBSNModule(clients["base"], cfg)
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("failed init module: %v", err)
	}

	modules.Collector = collector.NewCollector(clients["base"], modules.Dex.Uniswap, modules.LiquidPools.Aave, modules.LiquidPools.Moonwell)

	return &modules, nil
}

func initializeDexModules(client *ethClient.Client, cfg config.Config) (*DexModules, error) {
	v3dexesRouterABI, err := utils.ReadAbi(cfg.DexConfig.Pancake.RouterABIPath)
	if err != nil {
		return nil, err
	}

	v3dexesQuoterABI, err := utils.ReadAbi(cfg.DexConfig.Pancake.QuoterABIPath)
	if err != nil {
		return nil, err
	}

	pancake, err := dex.NewV3Router(client, common.HexToAddress(cfg.DexConfig.Pancake.RouterCA), common.HexToAddress(cfg.DexConfig.Pancake.QuoterCA), v3dexesRouterABI, v3dexesQuoterABI, cfg.DexConfig.Pancake.Fee, cfg.DexConfig.SqrtPriceLimitX96)
	if err != nil {
		return nil, fmt.Errorf("failed init Pancake: %v", err)
	}

	uniswap, err := dex.NewV3Router(client, common.HexToAddress(cfg.DexConfig.Uniswap.RouterCA), common.HexToAddress(cfg.DexConfig.Uniswap.QuoterCA), v3dexesRouterABI, v3dexesQuoterABI, cfg.DexConfig.Uniswap.Fee, cfg.DexConfig.SqrtPriceLimitX96)
	if err != nil {
		return nil, fmt.Errorf("failed init Uniswap: %v", err)
	}

	woofiABI, err := utils.ReadAbi(cfg.DexConfig.Woofi.ABIPath)
	if err != nil {
		return nil, err
	}

	woofi, err := dex.NewWooFi(client, common.HexToAddress(cfg.DexConfig.Woofi.CA), woofiABI)
	if err != nil {
		return nil, fmt.Errorf("failed init Woofi: %v", err)
	}

	return &DexModules{
		Pancake: pancake,
		Uniswap: uniswap,
		Woofi:   woofi,
	}, nil
}

func initializeRefuel(clients map[string]*ethClient.Client, cfg config.Config) (*refuel.Refuel, error) {
	abi, err := utils.ReadAbi(cfg.RefuelConfig.ABIPath)
	if err != nil {
		return nil, err
	}

	refuel_addresses := map[string]common.Address{
		"optimism":  common.HexToAddress(cfg.RefuelConfig.OptimismSocket),
		"arbitrum":  common.HexToAddress(cfg.RefuelConfig.ArbitrumSocket),
		"avalanche": common.HexToAddress(cfg.RefuelConfig.AvalancheSocket),
		"polygon":   common.HexToAddress(cfg.RefuelConfig.PolygonSocket),
		"base":      common.HexToAddress(cfg.RefuelConfig.BaseSocket),
	}

	refuel_chaind_ids := map[string]*big.Int{
		"optimism":  big.NewInt(10),
		"arbitrum":  big.NewInt(42161),
		"avalanche": big.NewInt(43114),
		"polygon":   big.NewInt(137),
		"base":      big.NewInt(8453),
	}

	return refuel.NewRefuel(clients, refuel_addresses, refuel_chaind_ids, abi)
}

func initializeStargate(clients map[string]*ethClient.Client, cfg config.Config) (*bridge.Stargate, error) {
	swap_abi, err := utils.ReadAbi(cfg.BridgeConfig.SwapABIPath)
	if err != nil {
		return nil, err
	}

	fee_abi, err := utils.ReadAbi(cfg.BridgeConfig.FeeABIPath)
	if err != nil {
		return nil, err
	}

	swap_cas := map[string]common.Address{
		"ethereum":  common.HexToAddress(cfg.BridgeConfig.SwapAddresses["ethereum"]),
		"bsc":       common.HexToAddress(cfg.BridgeConfig.SwapAddresses["bsc"]),
		"avalanche": common.HexToAddress(cfg.BridgeConfig.SwapAddresses["avalanche"]),
		"polygon":   common.HexToAddress(cfg.BridgeConfig.SwapAddresses["polygon"]),
		"arbitrum":  common.HexToAddress(cfg.BridgeConfig.SwapAddresses["arbitrum"]),
		"optimism":  common.HexToAddress(cfg.BridgeConfig.SwapAddresses["optimism"]),
	}

	fee_cas := map[string]common.Address{
		"ethereum":  common.HexToAddress(cfg.BridgeConfig.FeeAdresses["ethereum"]),
		"bsc":       common.HexToAddress(cfg.BridgeConfig.FeeAdresses["bsc"]),
		"avalanche": common.HexToAddress(cfg.BridgeConfig.FeeAdresses["avalanche"]),
		"polygon":   common.HexToAddress(cfg.BridgeConfig.FeeAdresses["polygon"]),
		"arbitrum":  common.HexToAddress(cfg.BridgeConfig.FeeAdresses["arbitrum"]),
		"optimism":  common.HexToAddress(cfg.BridgeConfig.FeeAdresses["optimism"]),
	}

	pool_ids := map[string]*big.Int{
		"eth_usdc": big.NewInt(1),
		"eth_usdt": big.NewInt(2),
		"eth_dai":  big.NewInt(3),
		"eth_eth":  big.NewInt(13),

		"bsc_usdt": big.NewInt(2),

		"avalanche_usdc": big.NewInt(1),
		"avalanche_usdt": big.NewInt(2),

		"polygon_usdc": big.NewInt(1),
		"polygon_usdt": big.NewInt(2),

		"arbitrum_usdc": big.NewInt(1),
		"arbitrum_usdt": big.NewInt(2),
		"arbitrum_eth":  big.NewInt(13),

		"optimism_usdc": big.NewInt(1),
		"optimism_dai":  big.NewInt(3),
		"optimism_eth":  big.NewInt(13),

		"base_usdc": big.NewInt(1),
		"base_eth":  big.NewInt(13),
	}

	chain_ids := map[string]uint16{
		"ethereum":  101,
		"bnb":       102,
		"avalanche": 106,
		"polygon":   109,
		"arbitrum":  110,
		"optimism":  111,
		"fantom":    112,
		"base":      184,
		"linea":     183,
	}
	return bridge.NewStargate(clients, swap_cas, fee_cas, pool_ids, chain_ids, swap_abi, fee_abi)
}

func initializeLiquidPoolsModules(client *ethClient.Client, cfg config.Config) (*LiquidPoolsModules, error) {
	aaveABI, err := utils.ReadAbi(cfg.LiquidPoolsConfig.Aave.ABIPath)
	if err != nil {
		return nil, err
	}

	aave, err := aave.NewAave(client, common.HexToAddress(cfg.LiquidPoolsConfig.Aave.ProxyBase), common.HexToAddress(cfg.LiquidPoolsConfig.Aave.EthPool), aaveABI)
	if err != nil {
		return nil, fmt.Errorf("failed init Aave: %v", err)
	}

	moonwellABI, err := utils.ReadAbi(cfg.LiquidPoolsConfig.Moonwell.ABIPath)
	if err != nil {
		return nil, err
	}
	moonwellWETHAbi, err := utils.ReadAbi(cfg.LiquidPoolsConfig.Moonwell.MWethABIPath)
	if err != nil {
		return nil, err
	}

	moonwell, err := moonwell.NewMoonwell(client, common.HexToAddress(cfg.LiquidPoolsConfig.Moonwell.CA), common.HexToAddress(cfg.LiquidPoolsConfig.Moonwell.METHCA), moonwellABI, moonwellWETHAbi)
	if err != nil {
		return nil, fmt.Errorf("failed init Moonwell: %v", err)
	}

	return &LiquidPoolsModules{
		Aave:     aave,
		Moonwell: moonwell,
	}, nil
}

func initializeNFTMintsModules(client *ethClient.Client, cfg config.Config) (*NFTMintsModules, error) {
	zoraABI, err := utils.ReadAbi(cfg.NFTMintsConfig.Zora.ABIPath)
	if err != nil {
		return nil, err
	}
	zora, err := nftmints.NewZora(client, common.HexToAddress(cfg.NFTMintsConfig.Zora.CA), zoraABI)
	if err != nil {
		return nil, fmt.Errorf("failed init Zora: %v", err)
	}

	nft2meABI, err := utils.ReadAbi(cfg.NFTMintsConfig.NFT2Me.ABIPath)
	if err != nil {
		return nil, err
	}

	nft2me, err := nftmints.NewNft2Me(client, nft2meABI)
	if err != nil {
		return nil, fmt.Errorf("failed init NFT2Me: %v", err)
	}

	return &NFTMintsModules{
		Zora:   zora,
		NFT2Me: nft2me,
	}, nil
}

func initializeBSNModule(client *ethClient.Client, cfg config.Config) (*domains.BSN, error) {
	regABI, err := utils.ReadAbi(cfg.DomainsConfig.RegisterABIPath)
	if err != nil {
		return nil, err
	}
	resABI, err := utils.ReadAbi(cfg.DomainsConfig.ResolverABIPath)
	if err != nil {
		return nil, err
	}

	return domains.NewBSN(client, common.HexToAddress(cfg.DomainsConfig.RegisterCA), common.HexToAddress(cfg.DomainsConfig.ResolverCA), regABI, resABI)
}
