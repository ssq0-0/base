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
	"fmt"

	"golang.org/x/sync/errgroup"
)

type Modules struct {
	Dex         DexModules
	Bridge      *bridge.Stargate
	Dmail       *dmail.Dmail
	Domains     *domains.BSN
	LiquidPools LiquidPoolsModules
	NFTMints    NFTMintsModules
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
		modules.Bridge, err = bridge.NewStargate(clients, config.LZ_Main_CA, cfg.BridgeConfig.SwapABIPath, cfg.BridgeConfig.FeeABIPath)
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
		modules.Domains, err = domains.NewBSN(clients["base"], cfg.DomainsConfig.RegisterCA, cfg.DomainsConfig.ResolverCA, cfg.DomainsConfig.RegisterABIPath, cfg.DomainsConfig.ResolverABIPath)
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("failed init module: %v", err)
	}

	modules.Collector = collector.NewCollector(clients["base"], modules.Dex.Uniswap, modules.LiquidPools.Aave, modules.LiquidPools.Moonwell)

	return &modules, nil
}

func initializeDexModules(client *ethClient.Client, cfg config.Config) (DexModules, error) {
	pancake, err := dex.NewV3Router(client, cfg.DexConfig.Pancake.RouterCA, cfg.DexConfig.Pancake.QuoterCA, cfg.DexConfig.Pancake.RouterABIPath, cfg.DexConfig.Pancake.QuoterABIPath, cfg.DexConfig.Pancake.Fee, cfg.DexConfig.SqrtPriceLimitX96)
	if err != nil {
		return DexModules{}, fmt.Errorf("failed init Pancake: %v", err)
	}

	uniswap, err := dex.NewV3Router(client, cfg.DexConfig.Uniswap.RouterCA, cfg.DexConfig.Uniswap.QuoterCA, cfg.DexConfig.Uniswap.RouterABIPath, cfg.DexConfig.Uniswap.QuoterABIPath, cfg.DexConfig.Uniswap.Fee, cfg.DexConfig.SqrtPriceLimitX96)
	if err != nil {
		return DexModules{}, fmt.Errorf("failed init Uniswap: %v", err)
	}

	woofi, err := dex.NewWooFi(client, cfg.DexConfig.Woofi.CA, cfg.DexConfig.Woofi.ABIPath)
	if err != nil {
		return DexModules{}, fmt.Errorf("failed init Woofi: %v", err)
	}

	return DexModules{
		Pancake: pancake,
		Uniswap: uniswap,
		Woofi:   woofi,
	}, nil
}

func initializeLiquidPoolsModules(client *ethClient.Client, cfg config.Config) (LiquidPoolsModules, error) {
	aave, err := aave.NewAave(client, cfg.LiquidPoolsConfig.Aave.ProxyBase, cfg.LiquidPoolsConfig.Aave.EthPool, cfg.LiquidPoolsConfig.Aave.ABIPath)
	if err != nil {
		return LiquidPoolsModules{}, fmt.Errorf("failed init Aave: %v", err)
	}

	moonwell, err := moonwell.NewMoonwell(client, cfg.LiquidPoolsConfig.Moonwell.CA, cfg.LiquidPoolsConfig.Moonwell.METHCA, cfg.LiquidPoolsConfig.Moonwell.ABIPath, cfg.LiquidPoolsConfig.Moonwell.MWethABIPath)
	if err != nil {
		return LiquidPoolsModules{}, fmt.Errorf("failed init Moonwell: %v", err)
	}

	return LiquidPoolsModules{
		Aave:     aave,
		Moonwell: moonwell,
	}, nil
}

func initializeNFTMintsModules(client *ethClient.Client, cfg config.Config) (NFTMintsModules, error) {
	zora, err := nftmints.NewZora(client, cfg.NFTMintsConfig.Zora.CA, cfg.NFTMintsConfig.Zora.ABIPath)
	if err != nil {
		return NFTMintsModules{}, fmt.Errorf("failed init Zora: %v", err)
	}

	nft2me, err := nftmints.NewNft2Me(client, cfg.NFTMintsConfig.NFT2Me.ABIPath)
	if err != nil {
		return NFTMintsModules{}, fmt.Errorf("failed init NFT2Me: %v", err)
	}

	return NFTMintsModules{
		Zora:   zora,
		NFT2Me: nft2me,
	}, nil
}
