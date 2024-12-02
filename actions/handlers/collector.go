package handlers

import (
	"base/account"
	"base/config"
	"base/ethClient"
	"base/modules"
)

type CollectorHandler struct{}

func (c *CollectorHandler) Execute(acc *account.Account, mods modules.Modules, client *ethClient.Client, config *config.Config) error {
	return mods.Collector.Collect(acc)
}
