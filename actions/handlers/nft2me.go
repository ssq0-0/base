package handlers

import (
	"base/account"
	"base/actions/types"
	"base/config"
	"base/ethClient"
	"base/modules"
	"math/big"
)

type Nft2MeHandler struct {
	NftMintParams types.NftMintParams
}

func (nh Nft2MeHandler) Execute(acc *account.Account, mods modules.Modules, client *ethClient.Client, config *config.Config) error {
	return mods.NFTMints.NFT2Me.Mint(nh.NftMintParams.MintCA, big.NewInt(1), nh.NftMintParams.Price, acc)
}
