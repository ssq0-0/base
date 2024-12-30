package handlers

import (
	"base/account"
	"base/actions/types"
	"base/config"
	"base/ethClient"
	"base/modules"
)

type BridgeHandler struct {
	BridgeParams types.BridgeParams
}

func (bh BridgeHandler) Execute(acc *account.Account, mods modules.Modules, client *ethClient.Client, config *config.Config) error {
	return mods.Bridge.SwapStable(bh.BridgeParams.FromChain, bh.BridgeParams.DstChain, bh.BridgeParams.Token, bh.BridgeParams.AmountToBridge, acc)
}
