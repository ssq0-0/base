package handlers

import (
	"base/account"
	"base/actions/types"
	"base/config"
	"base/ethClient"
	"base/modules"
)

type ZoraHandler struct {
	NftMintParams types.NftMintParams
}

func (zh ZoraHandler) Execute(acc *account.Account, mods modules.Modules, client *ethClient.Client, config *config.Config) error {
	return mods.NFTMints.Zora.Mint(zh.NftMintParams.MintCA, zh.NftMintParams.Price, acc)
}
