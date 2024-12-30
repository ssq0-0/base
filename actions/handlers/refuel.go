package handlers

import (
	"base/account"
	"base/actions/types"
	cfg "base/config"
	"base/ethClient"
	"base/modules"
)

type RefuelHandler struct {
	RefuelParams types.RefuelParams
}

func (rh *RefuelHandler) Execute(acc *account.Account, mods modules.Modules, client *ethClient.Client, config *cfg.Config) error {
	return mods.Refuel.Refuel(rh.RefuelParams.ScrChain, rh.RefuelParams.DstChain, acc)
}
